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
        
        s3_access_key_id: ${{ secrets.README_TEMPLATE_TOKEN }}
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

### Normal variables you can put into your template file

| Variable                 | Example |
| ------------------------ | ------- |
| {{ USERNAME }}            | probablykasper
| {{ NAME }}                | Kasper
| {{ EMAIL }}               | email@example.com
| {{ USER_ID }}             | MDQ6VXNlcjExMzE1NDky
| {{ BIO }}                 | Fullstack developer from Norway
| {{ COMPANY }}             | Microscopicsoft
| {{ LOCATION }}            | Norway
| {{ TWITTER_USERNAME }}    | probablykasper
| {{ AVATAR_URL }}          | https://avatars0.githubusercontent.com/u/11315492?u=c501da00e9b817ffc78faab6c630f236ac2738cf&v=4
| {{ WEBSITE_URL }}         | https://kasper.space/
| {{ SIGNUP_TIMESTAMP }}    | 2015-03-04T14:48:35Z
| {{ SIGNUP_DATE }}         | March 4th 2015
| {{ SIGNUP_DATE2 }}        | 2015-03-04
| {{ SIGNUP_YEAR }}         | 2015
| {{ SIGNUP_AGO }}          | 5 years ago
| {{ TOTAL_REPOS_SIZE_KB }} | 707453
| {{ TOTAL_REPOS_SIZE_MB }} | 707.5
| {{ TOTAL_REPOS_SIZE_GB }} | 0.71
| {{ TOTAL_REPOSITORIES }}  | 46
| {{ CURRENT_REPO_FULL_NAME }} | probablykasper/readme-template-action
