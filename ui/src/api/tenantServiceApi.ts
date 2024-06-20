import { TenantServiceApi } from './gen'
import config from '@/api/config.ts'

const tenantServiceApi = new TenantServiceApi(config)
export default tenantServiceApi
