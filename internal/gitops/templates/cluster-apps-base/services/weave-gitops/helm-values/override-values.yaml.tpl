adminUser:
  create: true
  username: admin
  passwordHash: {{ .Secrets.WeaveGitOps.PasswordHash }}
