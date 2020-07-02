# portal

## Setup

Install Vue CLI

```
npm install -g @vue/cli
```

Install required node_modules

```
npm install
```

Build static bundle

```
npm run build
```

Symlink dist
- This is a bit tricky. We need to link the built dist folder to a dist folder located in the backend repo. This location needs to be defined in your backend makefile.

Define dist location in backend makefile
```UI_DIR=./cmd/portal/dist```

Linux/macos
```
ln -s LOCATION_OF_BUILT_DIST UI_DIR
```

## Optional

## Launch Web GUI

```
vue ui
```

### Compiles and hot-reloads for development
```
npm run serve
```
