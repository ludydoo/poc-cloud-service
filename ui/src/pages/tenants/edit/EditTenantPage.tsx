import { Heading } from '@/components/heading.tsx'
import { useForm } from 'react-hook-form'
import { Field, FieldGroup, Fieldset, Label } from '@/components/fieldset.tsx'
import { Input } from '@/components/input.tsx'
import { Textarea } from '@/components/textarea.tsx'
import { Button } from '@/components/button.tsx'
import { useTenant, useUpdateTenant } from '@/api/queries.ts'
import { useCallback, useEffect } from 'react'
import { TenantServiceUpdateTenantBody } from '@/api'
import { parse, stringify } from 'yaml'
import { useNavigate, useParams } from 'react-router-dom'
import {
  defaultPath,
  defaultRepositoryURL,
  defaultTargetRevision,
} from '@/pages/tenants/constants.ts'
interface Data {
  id: string
  repoURL: string
  path: string
  targetRevision: string
  helmValues: string
}

export default function EditTenantPage() {
  const { id } = useParams<{ id: string }>()
  if (!id) {
    throw new Error('No tenant ID provided')
  }
  const { data, isLoading, isError } = useTenant(id)
  const { mutate } = useUpdateTenant()
  const { register, handleSubmit, reset } = useForm<Data>()
  useEffect(() => {
    if (!data) return
    reset({
      id: data.tenant.id,
      repoURL: data.tenant.source?.repoUrl || '',
      targetRevision:
        data.tenant.source?.targetRevision || defaultTargetRevision,
      helmValues: data.tenant.source?.helm?.values
        ? stringify(data.tenant.source?.helm?.values, null, 2)
        : '',
    })
  }, [data, reset])

  const nav = useNavigate()
  const onSubmit = useCallback(
    (data: Data) => {
      const req: { id: string } & TenantServiceUpdateTenantBody = {
        id,
        source: {
          repoUrl: data.repoURL,
          path: data.path,
          targetRevision: data.targetRevision,
          helm: {
            values: parse(data.helmValues) as object,
          },
        },
      }
      mutate(req, {
        onSuccess: data => {
          nav(`/tenants/${data.tenant.id}`)
        },
      })
    },
    [id, mutate, nav],
  )

  return (
    <form className="mx-auto max-w-2xl" onSubmit={handleSubmit(onSubmit)}>
      <Heading>Edit tenant {data ? data.tenant.id : ''}</Heading>
      <div className="space-y-8">
        <Fieldset className="mt-6">
          <FieldGroup>
            <Field>
              <Label>ID</Label>
              <Input className="font-mono" {...register('id')} readOnly />
            </Field>
            <Field>
              <Label>Repository URL</Label>
              <Input
                {...register('repoURL')}
                placeholder={defaultRepositoryURL}
              />
            </Field>
            <Field>
              <Label>Path</Label>
              <Input {...register('path')} placeholder={defaultPath} />
            </Field>
            <Field>
              <Label>Target Revision</Label>
              <Input
                {...register('targetRevision')}
                placeholder={defaultTargetRevision}
              />
            </Field>
            <Field>
              <Label>Helm Values</Label>
              <Textarea
                rows={10}
                className="font-mono"
                {...register('helmValues')}
              />
            </Field>
          </FieldGroup>
        </Fieldset>
        <Button type="submit">Update tenant</Button>
      </div>
    </form>
  )
}
