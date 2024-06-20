import { Heading } from '@/components/heading.tsx'
import {
  DescriptionDetails,
  DescriptionList,
  DescriptionTerm,
} from '@/components/description-list.tsx'
import { useTenant } from '@/api/queries.ts'
import { useParams } from 'react-router-dom'
import { Tenant } from '@/api'
import { defaultPath, defaultRepositoryURL } from '@/pages/tenants/constants.ts'
import clsx from 'clsx'
import { Button } from '@/components/button.tsx'
import { stringify } from 'yaml'

function TenantPath({ tenant }: { tenant: Tenant }) {
  const hasPath = !!tenant.source?.path
  const path = tenant.source?.path || `${defaultPath} (default)`
  return (
    <>
      <DescriptionTerm>Path</DescriptionTerm>
      <DescriptionDetails className={clsx(!hasPath && 'italic text-gray-400')}>
        {path}
      </DescriptionDetails>
    </>
  )
}

function TenantRepoUrl({ tenant }: { tenant: Tenant }) {
  const hasRepoUrl = !!tenant.source?.repoUrl
  const repoUrl = tenant.source?.repoUrl || `${defaultRepositoryURL} (default)`
  return (
    <>
      <DescriptionTerm>Repository URL</DescriptionTerm>
      <DescriptionDetails
        className={clsx(!hasRepoUrl && 'italic text-gray-400')}
      >
        {repoUrl}
      </DescriptionDetails>
    </>
  )
}

export default function TenantDetailsPage() {
  // eslint-disable-next-line @typescript-eslint/no-unsafe-call
  const { id } = useParams<{ id: string }>()
  if (!id) {
    throw new Error('No tenant ID provided')
  }
  const { data, isLoading, isError } = useTenant(id)
  return (
    <>
      <div className="flex flex-row items-center justify-between">
        <Heading>Tenant</Heading>
        <Button to={`/tenants/${id}/edit`} color="dark/zinc">
          Edit tenant
        </Button>
      </div>
      <div className="mt-6">
        <DescriptionList>
          <DescriptionTerm>ID</DescriptionTerm>
          <DescriptionDetails>{data && data.tenant.id}</DescriptionDetails>
          {data && <TenantRepoUrl tenant={data.tenant} />}
          {data && <TenantPath tenant={data.tenant} />}
          {data && (
            <>
              <DescriptionTerm>Helm values</DescriptionTerm>
              <DescriptionDetails>
                <pre>
                  {stringify(data.tenant.source?.helm?.values, null, 2)}
                </pre>
              </DescriptionDetails>
            </>
          )}
        </DescriptionList>
      </div>
    </>
  )
}
