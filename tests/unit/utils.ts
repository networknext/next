import { Wrapper } from '@vue/test-utils'
import { CombinedVueInstance } from 'vue/types/vue'

export function waitFor (
  wrapper:
    Wrapper<CombinedVueInstance<any, object, object, object, Record<never, any>>>,
  selector: string
): Promise<any> {
  return new Promise(resolve => {
    const timer = setInterval(() => {
      const el = wrapper.findAll(selector)
      if (el.length > 0) {
        clearInterval(timer)
        resolve()
      }
    }, 100)
  })
}
