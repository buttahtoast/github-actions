# S3 Mirror

Mirror binaries continously to S3 based on Github releases.

## Example usage

Check out [`example.yaml`](./example.yaml), that's how you can configure the binaries you would like to download.

Workflow:

```yaml
name: S3 Mirror
on:
  schedule:
    - cron: '0 1 * * *' # Every day at 1 am
  push:
    branches: [ main ]
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
      with:
        fetch-depth: 0
    - name: Generate README.md
      uses: buttahtoast/github-actions/s3-mirror@main
      with:
        
        s3_access_key: ${{ secrets.S3_ACCESS_KEY }}
        s3_secret_key: ${{ secrets.S3_SECRET_KEY }}
        template: TEMPLATE.md
        output: README.md
```

  s3_bucket:
    description: 'S3 bucket to upload binaries'
    required: true
  config:
    description: 'YAML config file for binaries to download'
    required: false
    default: 'config.yaml'
  s3_region:
    description: 'S3 Region'
    required: true
  s3_access_key_id:
    description: 'S3 Access key ID'
    required: true
  s3_secret_access_key:
    description: 'S3 Secret access key'
    required: true
  s3_endpoint:
    description: 'S3 endpoint'
    required: true
  s3_tlssecure:
    description

**config.yaml**:
```yaml
bins:
- name: "talos"
  versions:
    github: "https://github.com/siderolabs/talos"
    semver: ">=1.7.0"
  targets:
    - url: "{{.github}}/releases/download/{{.version}}/vmlinuz-{{.arch}}.xz"
      destination: "talos/vmlinuz-{{.version}}-{{.arch}}.xz"
    - url: "{{.github}}/releases/download/{{.version}}/initramfs-{{.arch}}.xz"
      destination: "talos/initramfs-{{.version}}-{{.arch}}.xz"
  arch: ["amd64", "arm64"]
  os: [ "linux" ]

- name: "cri-tools"
  versions:
    github: "https://github.com/kubernetes-sigs/cri-tools"
    semver: ">=1.25.0"
  targets:
    - url: "{{.github}}/releases/download/{{.version}}/crictl-{{.version}}-{{.os}}-{{.arch}}.tar.gz"
      checksum: "{{.github}}/releases/download/{{.version}}/crictl-{{.version}}-{{.os}}-{{.arch}}.tar.gz.sha256"
      destination: "crictl/crictl-{{.version}}-{{.os}}-{{.arch}}.tar.gz"
  arch: ["amd64", "arm64"]
  os: [ "linux" ]

- name: "kubernetes"
  versions:
    github: "https://github.com/kubernetes/kubernetes"
    semver: ">=1.25.0"
  targets:
    - url: "https://dl.k8s.io/{{.version}}/bin/{{.os}}/{{.arch}}/{{.bin}}"
      checksum: "https://dl.k8s.io/{{.version}}/bin/{{.os}}/{{.arch}}/{{.bin}}.sha256"
      destination: "kubernetes/{{.bin}}-{{.version}}-{{.os}}-{{.arch}}.tar.gz"
  arch: ["amd64", "arm64"]
  os: [ "linux" ]
  bins:
    - kubeadm
    - kubelet
    - kubectl
```


## Variables

| Action Variable                 | Environment Variable |
| ------------------------ | ------- |
| `s3_bucket`           | `S3_BUCKET` |
| `config`               | `CONFIG_FILE` |
| `s3_region`             | `AWS_REGION` |
| `s3_access_key`           | `AWS_ACCESS_KEY_ID` |
| `s3_secret_key`              | `AWS_SECRET_ACCESS_KEY` |
| `s3_endpoint`          | `S3_ENDPOINT` |
| `s3_tlssecure`          | `S3_TLSSECURE` |
| `log_level`   | `LOG_LEVEL` |