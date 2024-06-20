import { Heading, Subheading } from '@/components/heading.tsx'
import {
  DescriptionDetails,
  DescriptionList,
  DescriptionTerm,
} from '@/components/description-list.tsx'
import { useDeleteTenant, useTenant } from '@/api/queries.ts'
import { useNavigate, useParams } from 'react-router-dom'
import { Tenant } from '@/api'
import {
  defaultPath,
  defaultRepositoryURL,
  defaultTargetRevision,
} from '@/pages/tenants/constants.ts'
import clsx from 'clsx'
import { Button } from '@/components/button.tsx'
import { stringify } from 'yaml'
import { Text, TextLink } from '@/components/text.tsx'
import { ChevronRightIcon } from '@heroicons/react/24/outline'
import { useCallback, useState } from 'react'
import {
  Dialog,
  DialogActions,
  DialogBody,
  DialogTitle,
} from '@/components/dialog.tsx'

function TenantPath({ tenant }: { tenant: Tenant }) {
  const hasPath = !!tenant.source?.path
  const path = tenant.source?.path || `${defaultPath}`
  return (
    <>
      <DescriptionTerm>Path</DescriptionTerm>
      <DescriptionDetails className={clsx(!hasPath && 'text-gray-400')}>
        {path}
      </DescriptionDetails>
    </>
  )
}

function TenantRepoUrl({ tenant }: { tenant: Tenant }) {
  const hasRepoUrl = !!tenant.source?.repoUrl
  const repoUrl = tenant.source?.repoUrl || `${defaultRepositoryURL}`
  return (
    <>
      <DescriptionTerm>Repository URL</DescriptionTerm>
      <DescriptionDetails className={clsx(!hasRepoUrl && 'text-gray-400')}>
        {repoUrl}
      </DescriptionDetails>
    </>
  )
}

function TenantTargetRevision({ tenant }: { tenant: Tenant }) {
  const hasValue = !!tenant.source?.targetRevision
  const value = tenant.source?.targetRevision || `${defaultTargetRevision}`
  return (
    <>
      <DescriptionTerm>Target Revision</DescriptionTerm>
      <DescriptionDetails className={clsx(!hasValue && 'text-gray-400')}>
        {value}
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
  const [showConfirmation, onShowConfirmation] = useState(false)
  const { data, isLoading, isError } = useTenant(id)
  const { mutate: deleteTenant } = useDeleteTenant()
  const nav = useNavigate()
  const handleDeleteTenant = useCallback(() => {
    deleteTenant(id, {
      onSuccess: () => {
        nav('/tenants')
      },
    })
  }, [deleteTenant, id, nav])

  return (
    <div className="mx-auto max-w-3xl">
      <div className="flex flex-row items-center justify-between">
        <div className="flex flex-row items-center space-x-1">
          <TextLink to="/tenants">
            <Heading>Tenants</Heading>
          </TextLink>
          <ChevronRightIcon className="size-4" />
          <Heading>{data ? data.tenant.id : ''} </Heading>
        </div>

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
          {data && <TenantTargetRevision tenant={data.tenant} />}
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
      <div className="mt-6">
        <div className="space-y-4 rounded-lg border border-red-500 bg-red-50 p-4">
          <Subheading>
            <span className="text-red-500">Danger zone</span>
          </Subheading>
          <Button color="red" onClick={() => onShowConfirmation(true)}>
            Delete tenant
          </Button>
        </div>
      </div>
      <Dialog open={showConfirmation} onClose={onShowConfirmation}>
        <DialogTitle>Are you sure?</DialogTitle>
        <DialogBody>
          <Text>
            Are you sure you want to delete the tenant? This action cannot be
            undone.
          </Text>
        </DialogBody>
        <DialogActions>
          <Button color="dark/zinc" onClick={() => onShowConfirmation(false)}>
            Do not delete
          </Button>
          <Button color="red" onClick={handleDeleteTenant}>
            Yes, delete tenant {data ? data.tenant.id : ''}
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  )
}
