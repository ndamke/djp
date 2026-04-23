package store

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type FrontmatterDocument struct {
	Path string
	Root *yaml.Node
	Body string
}

func LoadFrontmatterDocument(path string) (*FrontmatterDocument, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("missing frontmatter start delimiter")
	}

	end := strings.Index(content[4:], "\n---\n")
	if end < 0 {
		return nil, fmt.Errorf("missing frontmatter end delimiter")
	}

	frontmatterText := content[4 : 4+end]
	body := content[4+end+5:]

	var root yaml.Node
	if err := yaml.Unmarshal([]byte(frontmatterText), &root); err != nil {
		return nil, err
	}

	return &FrontmatterDocument{
		Path: path,
		Root: &root,
		Body: body,
	}, nil
}

func (d *FrontmatterDocument) GetString(key string) string {
	node := d.lookupValue(key)
	if node == nil {
		return ""
	}

	return node.Value
}

func (d *FrontmatterDocument) SetString(key string, value string) {
	d.setScalar(key, "!!str", value)
}

func (d *FrontmatterDocument) SetInt(key string, value int) {
	d.setScalar(key, "!!int", fmt.Sprintf("%d", value))
}

func (d *FrontmatterDocument) Render() ([]byte, error) {
	payload, err := yaml.Marshal(d.Root)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteString("---\n")
	buf.Write(payload)
	buf.WriteString("---\n")
	buf.WriteString(d.Body)

	return buf.Bytes(), nil
}

func (d *FrontmatterDocument) setScalar(key string, tag string, value string) {
	mapping := d.rootMapping()
	if mapping == nil {
		return
	}

	if idx := d.lookupKeyIndex(key); idx >= 0 {
		mapping.Content[idx+1] = &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value}
		return
	}

	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value},
	)
}

func (d *FrontmatterDocument) lookupValue(key string) *yaml.Node {
	mapping := d.rootMapping()
	if mapping == nil {
		return nil
	}

	if idx := d.lookupKeyIndex(key); idx >= 0 {
		return mapping.Content[idx+1]
	}

	return nil
}

func (d *FrontmatterDocument) lookupKeyIndex(key string) int {
	mapping := d.rootMapping()
	if mapping == nil {
		return -1
	}

	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			return i
		}
	}

	return -1
}

func (d *FrontmatterDocument) rootMapping() *yaml.Node {
	if d.Root == nil {
		return nil
	}

	if d.Root.Kind == yaml.MappingNode {
		return d.Root
	}

	if d.Root.Kind == yaml.DocumentNode && len(d.Root.Content) > 0 {
		return d.Root.Content[0]
	}

	return nil
}
