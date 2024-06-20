import { Configuration } from '@/api/gen'

function getBasePath(): string {
  let path = import.meta.env.VITE_API_BASE_PATH
  if (!path) {
    path =
      window.location.protocol +
      '//' +
      window.location.hostname +
      (window.location.port ? ':' + window.location.port : '')
  }
  return path
}

const config = new Configuration({
  basePath: getBasePath(),
})

export default config
