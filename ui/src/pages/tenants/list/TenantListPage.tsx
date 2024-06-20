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

export default function TenantListPage() {
  const { data, isLoading, isError } = useTenants()

  return (
    <>
      <div className="flex flex-row items-center justify-between">
        <Heading>Tenants</Heading>
        <Button to="/tenants/create" color="dark/zinc">
          Create tenant
        </Button>
      </div>
      <Table>
        <TableHead>
          <TableRow>
            <TableHeader>ID</TableHeader>
          </TableRow>
        </TableHead>
        <TableBody>
          {isError && (
            <TableRow>
              <TableCell>Failed to load tenants</TableCell>
            </TableRow>
          )}
          {isLoading && (
            <TableRow>
              <TableCell>Loading</TableCell>
            </TableRow>
          )}
          {data &&
            data.tenants.map((tenant, index) => (
              <TableRow to={`/tenants/${tenant.id}`} key={index}>
                <TableCell>{tenant.id}</TableCell>
              </TableRow>
            ))}
          {data && data.tenants.length === 0 && (
            <TableRow>
              <TableCell>No tenants found</TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </>
  )
}
