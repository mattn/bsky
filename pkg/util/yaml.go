package util

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ReadYaml reads the specified path and returns an RNode.
// This is useful for filtering by KRM type.
func ReadYaml(path string) ([]*yaml.RNode, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading path %v", path)
	}

	input := bytes.NewReader(data)
	reader := kio.ByteReader{
		Reader: input,
		// TODO(jeremy): Do we want to exclude them?
		OmitReaderAnnotations: true,
	}

	nodes, err := reader.Read()
	if err != nil {
		return nil, errors.Wrapf(err, "Error unmarshaling %v", path)
	}

	return nodes, nil
}

// FindYamlFiles locates all the YAML files in some root.
// symlinks are evaluated
// Files are deduped (e.g. a symlink and its source will not be included twice if they are both in root).
func FindYamlFiles(root string) ([]string, error) {
	log := zapr.NewLogger(zap.L())

	paths := map[string]bool{}

	if _, err := os.Stat(root); err != nil && os.IsNotExist(err) {
		return []string{}, fmt.Errorf("FindYamlFiles invoked for non-existent path: %v", root)
	}

	// Walk the directory and add all YAML files.
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, walkErr error) error {
			// Skip non YAML files
			ext := strings.ToLower(filepath.Ext(info.Name()))

			if ext != ".yaml" && ext != ".yml" {
				return nil
			}
			p, err := filepath.EvalSymlinks(path)
			if err != nil {
				log.Error(err, "Failed to evaluate symlink", "path", path)
				return err
			}
			paths[p] = true
			return nil
		})

	results := []string{}
	for p := range paths {
		results = append(results, p)
	}
	return results, err
}
