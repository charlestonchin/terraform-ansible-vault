package main

import (
	_ "log"
  _ "fmt"
  "github.com/hashicorp/terraform/plugin"
  "github.com/ansible_vault/terraform-provider-ansible-vault/vault"
)

/*func main() {
  // Encrypt secret data
  str, err := vault.Encrypt("secret", "password")
  _ = str
  _ = err
  fmt.Printf(str)
  if err != nil {
    log.Fatal("Error:", err)
  }
  // Decrypt secret data
  //str, err := vault.Decrypt("secret", "password")

  // Write secret data to file
  //err := vault.EncryptFile("path/to/secret/file", "secret", "password")

  // Read existing secret
  //str, err := vault.DecryptFile("path/to/secret/file", "password")
}
*/
func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vault.Provider})
}
