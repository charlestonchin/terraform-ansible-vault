package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/hashicorp/hil"
	_ "github.com/hashicorp/hil/ast"
	_ "github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/helper/pathorcontents"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceFile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFileRead,

		Schema: map[string]*schema.Schema{
			"vault_string": {
                Optional:    true,
				Type:        schema.TypeString,
                Default:     "",
				Description: "Encrypted string",
				ConflictsWith: []string{"vault_file"},
			},
			"vault_password_file": {
                Optional:    true,
				Type:        schema.TypeString,
                Default:     "",
				Description: "Vault password file",	
				ConflictsWith: []string{"prompt_vault_pass"},
			},
			"prompt_vault_pass": {
                Optional:    true,
				Type:        schema.TypeString,
                Default:     "false",
				Description: "User input for vault password",	
				ConflictsWith: []string{"vault_password_file"},
				Sensitive:   true,
			},  
			"vault_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "file to read template from",
				ConflictsWith: []string{"vault_string"},
				// Make a "best effort" attempt to relativize the file path.
				StateFunc: func(v interface{}) string {
					if v == nil || v.(string) == "" {
						return ""
					}
					pwd, err := os.Getwd()
					if err != nil {
						return v.(string)
					}
					rel, err := filepath.Rel(pwd, v.(string))
					if err != nil {
						return v.(string)
					}
					return rel
				},
				Deprecated:    "Use the 'template' attribute instead.",
			},
		},
	}
}
func dataSourceFileRead(d *schema.ResourceData, meta interface{}) error {
	rendered, err := renderFile(d)
	if err != nil {
		return err
	}
	d.Set("rendered", rendered)
	d.SetId(hash(rendered))
	return nil
}

type VaultFileError error

func renderFile(d *schema.ResourceData) (string, error) {
	contents := ""
	filename := d.Get("vault_file").(string)
	if filename != "" {
		data, _, err := pathorcontents.Read(filename)
		if err != nil {
			return "", err
		}

		contents = data
	}

	return contents, nil
}


// type templateRenderError error

// func renderFile(d *schema.ResourceData) (string, error) {
// 	template := d.Get("template").(string)
// 	filename := d.Get("filename").(string)
// 	vars := d.Get("vars").(map[string]interface{})

// 	contents := template
// 	if template == "" && filename != "" {
// 		data, _, err := pathorcontents.Read(filename)
// 		if err != nil {
// 			return "", err
// 		}

// 		contents = data
// 	}

// 	rendered, err := execute(contents, vars)
// 	if err != nil {
// 		return "", templateRenderError(
// 			fmt.Errorf("failed to render %v: %v", filename, err),
// 		)
// 	}

// 	return rendered, nil
// }

// execute parses and executes a template using vars.
// func execute(s string, vars map[string]interface{}) (string, error) {
// 	root, err := hil.Parse(s)
// 	if err != nil {
// 		return "", err
// 	}

// 	varmap := make(map[string]ast.Variable)
// 	for k, v := range vars {
// 		// As far as I can tell, v is always a string.
// 		// If it's not, tell the user gracefully.
// 		s, ok := v.(string)
// 		if !ok {
// 			return "", fmt.Errorf("unexpected type for variable %q: %T", k, v)
// 		}
// 		varmap[k] = ast.Variable{
// 			Value: s,
// 			Type:  ast.TypeString,
// 		}
// 	}

// 	cfg := hil.EvalConfig{
// 		GlobalScope: &ast.BasicScope{
// 			VarMap:  varmap,
// 			FuncMap: config.Funcs(),
// 		},
// 	}

// 	result, err := hil.Eval(root, &cfg)
// 	if err != nil {
// 		return "", err
// 	}
// 	if result.Type != hil.TypeString {
// 		return "", fmt.Errorf("unexpected output hil.Type: %v", result.Type)
// 	}

// 	return result.Value.(string), nil
// }

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func validateVarsAttribute(v interface{}, key string) (ws []string, es []error) {
	// vars can only be primitives right now
	var badVars []string
	for k, v := range v.(map[string]interface{}) {
		switch v.(type) {
		case []interface{}:
			badVars = append(badVars, fmt.Sprintf("%s (list)", k))
		case map[string]interface{}:
			badVars = append(badVars, fmt.Sprintf("%s (map)", k))
		}
	}
	if len(badVars) > 0 {
		es = append(es, fmt.Errorf(
			"%s: cannot contain non-primitives; bad keys: %s",
			key, strings.Join(badVars, ", ")))
	}
	return
}
