package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencenter-cloud/opencenter-cli/internal/sops"
	"github.com/opencenter-cloud/opencenter-cli/internal/util/crypto"
)

// executeSOPSGenerateKey generates a new Age key pair
func executeSOPSGenerateKey(ctx context.Context, out, errOut io.Writer, keyFile string, updateSOPS bool, dryRun bool) error {
	if dryRun {
		fmt.Fprintf(out, "🧪 DRY RUN: Age key generation simulation\n")
	} else {
		fmt.Fprintf(out, "🔑 Generating new Age key pair for SOPS encryption...\n")
	}

	// Set default key file path if not specified
	if keyFile == "" {
		homeDir, _ := os.UserHomeDir()
		keyFile = filepath.Join(homeDir, ".config", "sops", "age", "keys.txt")
	}

	fmt.Fprintf(out, "📁 Key file path: %s\n", keyFile)
	fmt.Fprintf(out, "🔧 Update SOPS config: %t\n", updateSOPS)

	if dryRun {
		fmt.Fprintf(out, "🔑 Would generate new Age key pair\n")
		fmt.Fprintf(out, "📄 Would save private key to specified path\n")
		fmt.Fprintf(out, "🔑 Would display public key for configuration\n")
		if updateSOPS {
			fmt.Fprintf(out, "⚙️  Would update .sops.yaml configuration\n")
		}
		return nil
	}

	// Initialize key manager for the directory containing the key file
	keyDir := filepath.Dir(keyFile)
	km := sops.NewKeyManager(keyDir)

	// Generate new key pair
	keyPair, err := km.GenerateAgeKey()
	if err != nil {
		return fmt.Errorf("failed to generate Age key pair: %w", err)
	}

	// Check if key file already exists and backup if needed
	if _, err := os.Stat(keyFile); err == nil {
		fmt.Fprintf(errOut, "⚠️  Key file '%s' already exists, creating backup...\n", keyFile)
		backupPath := fmt.Sprintf("%s.backup-%s", keyFile, time.Now().Format("20060102-150405"))
		if err := copyFile(keyFile, backupPath); err == nil {
			fmt.Fprintf(out, "✅ Existing key backed up as: %s\n", backupPath)
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(keyDir, 0o700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Write the private key to the specified path
	// Format: # created: <timestamp>\n# public key: <public_key>\n<private_key>
	keyContent := fmt.Sprintf("# created: %s\n# public key: %s\n%s",
		time.Now().UTC().Format(time.RFC3339),
		keyPair.PublicKey,
		keyPair.PrivateKey)
	if !strings.HasSuffix(keyContent, "\n") {
		keyContent += "\n"
	}
	if err := os.WriteFile(keyFile, []byte(keyContent), 0o600); err != nil {
		return fmt.Errorf("failed to save Age key: %w", err)
	}

	fmt.Fprintf(out, "✅ Age key pair generated successfully!\n")
	fmt.Fprintf(out, "📁 Private key: %s\n", keyFile)
	fmt.Fprintf(out, "🔑 Public key: %s\n", keyPair.PublicKey)

	// Update SOPS configuration if requested
	if updateSOPS {
		if err := updateSOPSConfig(keyPair.PublicKey); err != nil {
			fmt.Fprintf(errOut, "⚠️  Failed to update .sops.yaml: %v\n", err)
			fmt.Fprintf(out, "💡 Please update .sops.yaml manually with the public key above\n")
		} else {
			fmt.Fprintf(out, "✅ Updated .sops.yaml configuration\n")
		}
	}

	return nil
}

// executeSOPSRotateKey rotates Age keys and re-encrypts secrets
func executeSOPSRotateKey(ctx context.Context, out, errOut io.Writer, keyFile, searchPath string, dryRun bool) error {
	fmt.Fprintf(out, "🔄 Starting Age key rotation...\n")

	// Initialize key manager
	homeDir, _ := os.UserHomeDir()
	keyDir := filepath.Join(homeDir, ".config", "sops", "age")
	km := sops.NewKeyManager(keyDir)

	// Use default key name if not specified
	keyName := "keys"
	if keyFile != "" {
		keyName = filepath.Base(strings.TrimSuffix(keyFile, filepath.Ext(keyFile)))
	}

	if dryRun {
		fmt.Fprintf(out, "🧪 DRY RUN: Age key rotation simulation\n")
		fmt.Fprintf(out, "📁 Key name: %s\n", keyName)
		fmt.Fprintf(out, "🔍 Search path: %s\n", searchPath)

		// Find SOPS files that would be re-encrypted
		fmt.Fprintf(out, "🔍 Searching for SOPS-encrypted files...\n")
		encryptor := sops.NewDefaultEncryptor(nil, nil)

		err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip excluded directories
			if info.IsDir() && shouldSkipDirectory(path) {
				return filepath.SkipDir
			}

			if !info.IsDir() && isSOPSYAMLFile(path) {
				if isEncrypted, err := encryptor.IsFileEncrypted(path); err == nil && isEncrypted {
					fmt.Fprintf(out, "  📄 Would re-encrypt: %s\n", path)
				}
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to search for encrypted files: %w", err)
		}

		return nil
	}

	fmt.Fprintf(out, "📁 Key name: %s\n", keyName)
	fmt.Fprintf(out, "🔍 Search path: %s\n", searchPath)

	// Load old key for backup
	oldKey, err := km.LoadAgeKey(keyName)
	if err != nil {
		return fmt.Errorf("failed to load existing key: %w", err)
	}

	// Create backup
	backupDir := filepath.Join(keyDir, "backups")
	if err := os.MkdirAll(backupDir, 0o700); err == nil {
		backupKM := sops.NewKeyManager(backupDir)
		backupName := fmt.Sprintf("%s-backup-%s", keyName, time.Now().Format("20060102-150405"))
		if err := backupKM.SaveAgeKey(oldKey, backupName); err == nil {
			fmt.Fprintf(out, "✅ Old key backed up as: %s\n", backupName)
		}
	}

	// Generate new key
	newKey, err := km.GenerateAgeKey()
	if err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	// Save new key
	if err := km.SaveAgeKey(newKey, keyName); err != nil {
		return fmt.Errorf("failed to save new key: %w", err)
	}

	// Re-encrypt files with new key
	encryptor := sops.NewDefaultEncryptor([]string{newKey.PublicKey}, nil)

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if info.IsDir() && shouldSkipDirectory(path) {
			return filepath.SkipDir
		}

		if !info.IsDir() && isSOPSYAMLFile(path) {
			if isEncrypted, err := encryptor.IsFileEncrypted(path); err == nil && isEncrypted {
				fmt.Fprintf(out, "🔄 Re-encrypting: %s\n", path)
				if err := encryptor.RotateKeys(ctx, path, []string{newKey.PublicKey}, nil); err != nil {
					fmt.Fprintf(errOut, "⚠️  Failed to re-encrypt %s: %v\n", path, err)
				}
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to re-encrypt files: %w", err)
	}

	fmt.Fprintf(out, "✅ Age key rotation completed successfully!\n")
	fmt.Fprintf(out, "🔑 New public key: %s\n", newKey.PublicKey)

	// Update SOPS configuration
	if err := updateSOPSConfig(newKey.PublicKey); err != nil {
		fmt.Fprintf(errOut, "⚠️  Failed to update .sops.yaml: %v\n", err)
		fmt.Fprintf(out, "💡 Please update .sops.yaml manually with the new public key\n")
	} else {
		fmt.Fprintf(out, "✅ Updated .sops.yaml configuration\n")
	}

	return nil
}

// executeSOPSBackupKey creates a backup of Age keys
func executeSOPSBackupKey(ctx context.Context, out, errOut io.Writer, keyFile, backupDir string, dryRun bool) error {
	if dryRun {
		fmt.Fprintf(out, "🧪 DRY RUN: Age key backup simulation\n")
	} else {
		fmt.Fprintf(out, "💾 Creating Age key backup...\n")
	}

	// Initialize key manager
	homeDir, _ := os.UserHomeDir()
	keyDir := filepath.Join(homeDir, ".config", "sops", "age")
	km := sops.NewKeyManager(keyDir)

	// Use default backup directory if not specified
	if backupDir == "" {
		backupDir = filepath.Join(keyDir, "backups")
	}

	fmt.Fprintf(out, "📁 Backup directory: %s\n", backupDir)

	if dryRun {
		fmt.Fprintf(out, "💾 Would create backup of all Age keys\n")
		fmt.Fprintf(out, "📄 Would backup .sops.yaml configuration if it exists\n")
		return nil
	}

	// Create backup
	if err := km.BackupKeys(backupDir); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	fmt.Fprintf(out, "✅ Age key backup created successfully!\n")
	fmt.Fprintf(out, "📁 Backup directory: %s\n", backupDir)

	// Also backup SOPS configuration if it exists
	if _, err := os.Stat(".sops.yaml"); err == nil {
		configBackup := filepath.Join(backupDir, fmt.Sprintf("sops-config-%s.yaml",
			time.Now().Format("20060102-150405")))

		if err := copyFile(".sops.yaml", configBackup); err != nil {
			fmt.Fprintf(errOut, "⚠️  Failed to backup SOPS configuration: %v\n", err)
		} else {
			fmt.Fprintf(out, "✅ SOPS configuration backed up to: %s\n", configBackup)
		}
	}

	return nil
}

// executeSOPSValidate validates Age key configuration
func executeSOPSValidate(ctx context.Context, out, errOut io.Writer, keyFile, configFile string, dryRun bool) error {
	// Check if debug mode is enabled
	debugMode := os.Getenv("OPENCENTER_DEBUG") != ""

	if dryRun {
		fmt.Fprintf(out, "🧪 DRY RUN: SOPS validation simulation\n")
	} else {
		fmt.Fprintf(out, "🔍 Validating SOPS configuration...\n")
	}

	if debugMode {
		fmt.Fprintf(out, "🐛 DEBUG MODE ENABLED\n")
		fmt.Fprintf(out, "🐛 Environment: OPENCENTER_DEBUG=%s\n", os.Getenv("OPENCENTER_DEBUG"))
	}

	// Initialize key manager and determine key path
	homeDir, _ := os.UserHomeDir()
	var keyPath string
	var keyDir string
	var keyName string

	if keyFile != "" {
		// Use the provided key file path directly
		if strings.HasPrefix(keyFile, "~") {
			keyPath = filepath.Join(homeDir, keyFile[1:])
		} else if filepath.IsAbs(keyFile) {
			keyPath = keyFile
		} else {
			// Relative path - make it absolute
			if absPath, err := filepath.Abs(keyFile); err == nil {
				keyPath = absPath
			} else {
				keyPath = keyFile
			}
		}
		keyDir = filepath.Dir(keyPath)
		keyName = filepath.Base(strings.TrimSuffix(keyPath, filepath.Ext(keyPath)))

		if debugMode {
			fmt.Fprintf(out, "🐛 Using custom key file: %s\n", keyFile)
			fmt.Fprintf(out, "🐛 Resolved key path: %s\n", keyPath)
		}
	} else {
		// Use default key location
		keyDir = filepath.Join(homeDir, ".config", "sops", "age")
		keyName = "keys"
		keyPath = filepath.Join(keyDir, fmt.Sprintf("%s.txt", keyName))
	}

	km := sops.NewKeyManager(keyDir)

	if debugMode {
		fmt.Fprintf(out, "🐛 Home directory: %s\n", homeDir)
		fmt.Fprintf(out, "🐛 Key directory: %s\n", keyDir)
	}

	fmt.Fprintf(out, "📁 Key name: %s\n", keyName)
	fmt.Fprintf(out, "📁 Key path: %s\n", keyPath)
	fmt.Fprintf(out, "📄 Config file: %s\n", configFile)

	if dryRun {
		fmt.Fprintf(out, "🔍 Would validate Age key file existence\n")
		fmt.Fprintf(out, "🔍 Would validate Age key format\n")
		fmt.Fprintf(out, "🔍 Would validate SOPS configuration\n")
		fmt.Fprintf(out, "🔍 Would test key access\n")
		fmt.Fprintf(out, "🔍 Would check SOPS installation\n")
		return nil
	}

	// Check if key exists
	if debugMode {
		fmt.Fprintf(out, "🐛 Checking key path: %s\n", keyPath)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		if debugMode {
			fmt.Fprintf(out, "🐛 Key file not found at: %s\n", keyPath)
			// List available keys
			if entries, err := os.ReadDir(keyDir); err == nil {
				fmt.Fprintf(out, "🐛 Available files in key directory:\n")
				for _, entry := range entries {
					fmt.Fprintf(out, "🐛   - %s\n", entry.Name())
				}
			} else {
				fmt.Fprintf(out, "🐛 Failed to list key directory: %v\n", err)
			}
		}
		return fmt.Errorf("❌ Age key file not found: %s", keyPath)
	}

	if debugMode {
		// Show file permissions
		if info, err := os.Stat(keyPath); err == nil {
			fmt.Fprintf(out, "🐛 Key file permissions: %s\n", info.Mode())
			fmt.Fprintf(out, "🐛 Key file size: %d bytes\n", info.Size())
		}
	}

	// Check if public key file exists, if not create it from private key
	publicKeyPath := strings.TrimSuffix(keyPath, filepath.Ext(keyPath)) + ".pub"
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		if debugMode {
			fmt.Fprintf(out, "🐛 Public key file not found at: %s\n", publicKeyPath)
			fmt.Fprintf(out, "🐛 Generating public key from private key...\n")
		} else {
			fmt.Fprintf(out, "ℹ️  Public key file not found, generating from private key...\n")
		}

		// Read the private key
		privateKeyData, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("❌ Failed to read private key: %w", err)
		}

		// Extract the actual key (skip comments and empty lines)
		privateKeyStr := extractAgeKeyFromContent(string(privateKeyData))
		if privateKeyStr == "" {
			return fmt.Errorf("❌ No valid age key found in file")
		}

		// Parse the private key to extract the public key
		keyPair, err := crypto.ParseAgeKey(privateKeyStr)
		if err != nil {
			return fmt.Errorf("❌ Failed to parse private key: %w", err)
		}

		// Save the public key
		if err := os.WriteFile(publicKeyPath, []byte(keyPair.PublicKey), 0o644); err != nil {
			return fmt.Errorf("❌ Failed to save public key: %w", err)
		}

		fmt.Fprintf(out, "✅ Created public key file: %s\n", publicKeyPath)
		if debugMode {
			fmt.Fprintf(out, "🐛 Public key: %s\n", keyPair.PublicKey)
		}
	} else if debugMode {
		fmt.Fprintf(out, "🐛 Public key file exists at: %s\n", publicKeyPath)
	}

	// Load key to get public key
	if debugMode {
		fmt.Fprintf(out, "🐛 Loading Age key...\n")
	}
	keyPair, err := km.LoadAgeKey(keyName)
	if err != nil {
		if debugMode {
			fmt.Fprintf(out, "🐛 Failed to load key: %v\n", err)
		}
		return fmt.Errorf("❌ Failed to load Age key: %w", err)
	}

	if debugMode {
		fmt.Fprintf(out, "🐛 Private key length: %d characters\n", len(keyPair.PrivateKey))
		fmt.Fprintf(out, "🐛 Public key: %s\n", keyPair.PublicKey)
		fmt.Fprintf(out, "🐛 Public key length: %d characters\n", len(keyPair.PublicKey))
	}

	// Validate key format
	if debugMode {
		fmt.Fprintf(out, "🐛 Validating Age key format...\n")
	}
	if err := km.ValidateAgeKey(keyPair.PublicKey); err != nil {
		if debugMode {
			fmt.Fprintf(out, "🐛 Key validation failed: %v\n", err)
		}
		return fmt.Errorf("❌ Age key validation failed: %w", err)
	}

	fmt.Fprintf(out, "✅ Age key validation passed\n")
	fmt.Fprintf(out, "🔑 Public key: %s\n", keyPair.PublicKey)

	// Validate SOPS configuration if it exists
	if debugMode {
		fmt.Fprintf(out, "🐛 Checking for SOPS config at: %s\n", configFile)
	}

	if _, err := os.Stat(configFile); err == nil {
		fmt.Fprintf(out, "🔍 Validating SOPS configuration file...\n")

		// Basic validation - check if public key is in config
		content, err := os.ReadFile(configFile)
		if err != nil {
			if debugMode {
				fmt.Fprintf(out, "🐛 Failed to read config: %v\n", err)
			}
			return fmt.Errorf("failed to read SOPS config: %w", err)
		}

		if debugMode {
			fmt.Fprintf(out, "🐛 Config file size: %d bytes\n", len(content))
			fmt.Fprintf(out, "🐛 Config file contents:\n")
			fmt.Fprintf(out, "--- BEGIN CONFIG ---\n")
			fmt.Fprintf(out, "%s\n", string(content))
			fmt.Fprintf(out, "--- END CONFIG ---\n")
		}

		if !strings.Contains(string(content), keyPair.PublicKey) {
			fmt.Fprintf(errOut, "⚠️  SOPS configuration does not contain current Age public key\n")
			fmt.Fprintf(out, "💡 Consider updating %s with the current public key\n", configFile)

			if debugMode {
				fmt.Fprintf(out, "🐛 Expected public key: %s\n", keyPair.PublicKey)
				// Show what age keys are in the config
				lines := strings.Split(string(content), "\n")
				fmt.Fprintf(out, "🐛 Age keys found in config:\n")
				for _, line := range lines {
					if strings.Contains(line, "age:") || strings.Contains(line, "age1") {
						fmt.Fprintf(out, "🐛   %s\n", strings.TrimSpace(line))
					}
				}
			}
		} else {
			fmt.Fprintf(out, "✅ SOPS configuration contains current Age public key\n")
		}
	} else {
		fmt.Fprintf(errOut, "⚠️  SOPS configuration file not found: %s\n", configFile)
		if debugMode {
			fmt.Fprintf(out, "🐛 Config file error: %v\n", err)
			// Check current directory
			if cwd, err := os.Getwd(); err == nil {
				fmt.Fprintf(out, "🐛 Current working directory: %s\n", cwd)
			}
		}
	}

	// Test key access
	fmt.Fprintf(out, "🧪 Testing key access...\n")
	if debugMode {
		fmt.Fprintf(out, "🐛 Validating key access permissions...\n")
	}
	if err := km.ValidateKeyAccess(keyName); err != nil {
		if debugMode {
			fmt.Fprintf(out, "🐛 Key access validation failed: %v\n", err)
		}
		return fmt.Errorf("❌ Key access test failed: %w", err)
	}

	fmt.Fprintf(out, "✅ Key access test passed\n")

	// Check SOPS installation
	if debugMode {
		fmt.Fprintf(out, "🐛 Checking SOPS installation...\n")
		// Check PATH
		if path := os.Getenv("PATH"); path != "" {
			fmt.Fprintf(out, "🐛 PATH: %s\n", path)
		}
	}

	manager := sops.NewSOPSManager()
	if version, err := manager.CheckSOPSVersion(ctx); err != nil {
		fmt.Fprintf(errOut, "⚠️  SOPS not found or not executable: %v\n", err)
		if debugMode {
			fmt.Fprintf(out, "🐛 SOPS check error details: %v\n", err)
		}
	} else {
		fmt.Fprintf(out, "✅ SOPS is installed: %s\n", version)
	}

	// Additional debug checks
	if debugMode {
		fmt.Fprintf(out, "\n🐛 === ADDITIONAL DEBUG INFORMATION ===\n")

		// Check SOPS_AGE_KEY_FILE environment variable
		if sopsKeyFile := os.Getenv("SOPS_AGE_KEY_FILE"); sopsKeyFile != "" {
			fmt.Fprintf(out, "🐛 SOPS_AGE_KEY_FILE: %s\n", sopsKeyFile)
		} else {
			fmt.Fprintf(out, "🐛 SOPS_AGE_KEY_FILE: (not set)\n")
		}

		// List all age keys in the key directory
		fmt.Fprintf(out, "🐛 All Age keys in key directory:\n")
		if keyNames, err := km.ListAgeKeys(); err == nil {
			for _, name := range keyNames {
				fmt.Fprintf(out, "🐛   - %s\n", name)
				if kp, err := km.LoadAgeKey(name); err == nil {
					fmt.Fprintf(out, "🐛     Public key: %s\n", kp.PublicKey)
				}
			}
		} else {
			fmt.Fprintf(out, "🐛 Failed to list keys: %v\n", err)
		}

		fmt.Fprintf(out, "🐛 === END DEBUG INFORMATION ===\n")
	}

	fmt.Fprintf(out, "✅ All validations completed successfully!\n")

	return nil
}
