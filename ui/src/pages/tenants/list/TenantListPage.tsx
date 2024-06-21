import { Heading } from '@/components/heading.tsx'
import { useTenants } from '@/api/queries.ts'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/table.tsx'
import { Button } from '@/components/button.tsx'

import TenantStatus from '@/components/tenantstatus.tsx'

export default function TenantListPage() {
  const { data, isLoading, isError } = useTenants()

  return (
    <div className="mx-auto max-w-3xl">
      <div className="flex flex-row items-center justify-between">
        <Heading>Tenants</Heading>
        <Button to="/tenants/create" color="dark/zinc">
          Create tenant
        </Button>
      </div>
      <Table className="mt-6">
        <TableHead>
          <TableRow>
            <TableHeader>ID</TableHeader>
            <TableHeader>Status</TableHeader>
          </TableRow>
        </TableHead>
        <TableBody>
          {isError && (
            <TableRow>
              <TableCell colSpan={2}>Failed to load tenants</TableCell>
            </TableRow>
          )}
          {isLoading && (
            <TableRow>
              <TableCell colSpan={2}>Loading</TableCell>
            </TableRow>
          )}
          {data &&
            data.tenants.map((tenant, index) => (
              <TableRow to={`/tenants/${tenant.id}`} key={index}>
                <TableCell>{tenant.id}</TableCell>
                <TableCell>
                  <TenantStatus status={tenant.application?.health?.status} />
                </TableCell>
              </TableRow>
            ))}
          {data && data.tenants.length === 0 && (
            <TableRow>
              <TableCell colSpan={2}>No tenants found</TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  )
}
