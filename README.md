# helm-push

`helm-push` is a [helm](https://github.com/kubernetes/helm) plugin that allows you to push chart package to TencentHub

## Installation

Install the latest version:
```shell
$ helm plugin install https://github.com/imroc/helm-push
```

Install a specific version:
```shell
$ helm plugin install https://github.com/imroc/helm-push --version 0.2.0
```

## Quick start

```shell
# Add your chart repository
$ helm repo add repo-name https://hub.tencentyun.com/charts/yournamespace --username yourusername --password yourpassword

# Push a chart package to your repository
$ helm push chartname.tar.gz reponame

# Push a chart directory to your repository
$ helm push . reponame

# Update Helm cache
$ helm repo update

# Fetch the chart
$ helm fetch reponame/chartname
```