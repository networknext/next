import camelCase from 'lodash/camelCase'

const requireModules = require.context('.', false, /\.ts$/)
const modules: any = {}

requireModules.keys().forEach((filename: string) => {
  if (filename === './index.ts') return

  const moduleName = camelCase(filename.replace(/(\.\/|\.ts)/g, ''))

  modules[moduleName] = requireModules(filename).default
})
export default modules
