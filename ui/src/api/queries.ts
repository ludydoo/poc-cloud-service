import { createQueryKeys } from '@lukemorales/query-key-factory'
import tenantServiceApi from './tenantServiceApi.ts'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { CreateTenantRequest, TenantServiceUpdateTenantBody } from './gen'

const tenantQueries = createQueryKeys('tenants', {
  list: {
    queryKey: ['list'],
    queryFn: () => tenantServiceApi.tenantServiceListTenants(),
  },
  one: (id: string) => ({
    queryKey: ['one', id],
    queryFn: () => tenantServiceApi.tenantServiceGetTenant({ id }),
  }),
})

export function useTenant(id: string) {
  return useQuery({
    ...tenantQueries['one'](id),
    enabled: !!id,
  })
}

export function useTenants() {
  return useQuery(tenantQueries['list'])
}

export function useCreateTenant() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateTenantRequest) =>
      tenantServiceApi.tenantServiceCreateTenant({ body: data }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: tenantQueries['list'].queryKey,
      })
    },
  })
}

export function useUpdateTenant() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      id,
      ...body
    }: TenantServiceUpdateTenantBody & { id: string }) => {
      return tenantServiceApi.tenantServiceUpdateTenant({ id, body })
    },
    onSuccess: async data => {
      await queryClient.invalidateQueries({
        queryKey: tenantQueries['list'].queryKey,
      })
      const id = data.tenant?.id
      if (id) {
        await queryClient.invalidateQueries({
          queryKey: tenantQueries['one'](id).queryKey,
        })
      }
    },
  })
}

export function useDeleteTenant() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) =>
      tenantServiceApi.tenantServiceDeleteTenant({ id }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: tenantQueries['list'].queryKey,
      })
    },
  })
}

export { tenantQueries }
