# SSH-MESSER

## Install

```shell
# config homebrew tap
brew tap ole3021/ssh-messer
brew install ssh-messer
# upgrade
brew upgrade ssh-messer
```

## Build & Release

```shell
# Create Tag version and push
git tag v0.2.2
git push origin v0.2.2
```

Github Actions will automatically build and release to homebrew throw [homebrew-ssh-messer](https://github.com/ole3021/homebrew-ssh-messer).
