import { Configuration } from '@/api/gen'

const config = new Configuration({
  basePath: import.meta.env.VITE_API_BASE_PATH,
})

export default config
