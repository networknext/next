# Portal

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
Windows
???????????????

## Optional

## Launch Web GUI

```
vue ui
```

### Compiles and hot-reloads for development
This breaks all API calls because of CORS
```
npm run serve
```

### Tests
E2E

I know a lot of us are not GUI people but I highly recommend using ```vue ui```. Within it you can setup a dashboard with tasks for running the dev server, builds for dev/prod, or tests

The easiest way I have found to run the E2E tests is to launch Vue UI, serve the development server using the 'Serve' task and then launch Cypress using the 'test:e2e' task

This will start the dev server on 127.0.0.1:8080 and then attach the e2e tests.

The other option is to run ```npm run serve``` in a terminal window and ```npm run test:e2e``` in another. Doing it this way will also require params to be passed in that I need to find...

### Auto linting setup

<<<<<<< HEAD
VSCode settings.json
=======
VSCode settings.js
>>>>>>> Added stuff to README for dev env setup
```
  "go.formatTool": "goimports",
  "go.useLanguageServer": true,
  "window.zoomLevel": 1,
  "explorer.confirmDelete": false,
  "typescript.updateImportsOnFileMove.enabled": "always",
  "javascript.updateImportsOnFileMove.enabled": "always",
  "workbench.iconTheme": "vscode-simpler-icons",
  "editor.suggestSelection": "first",
  "explorer.confirmDragAndDrop": false,
  "editor.tabSize": 2,
  "editor.detectIndentation": false,
  "C_Cpp.updateChannel": "Insiders",
  "json.maxItemsComputed": 100000,
  "editor.formatOnPaste": true,
  "editor.formatOnSave": true,
  "eslint.format.enable": true,
  "eslint.run": "onSave",
  "javascript.format.insertSpaceAfterConstructor": true,
  "typescript.format.insertSpaceAfterConstructor": true,
  "javascript.format.insertSpaceBeforeFunctionParenthesis": true,
  "typescript.format.insertSpaceBeforeFunctionParenthesis": true,
```

VSCode Plugins:
- beautify
- eslint