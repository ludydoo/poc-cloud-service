import { Heading } from '@/components/heading.tsx'
import { useForm } from 'react-hook-form'
import { Field, FieldGroup, Fieldset, Label } from '@/components/fieldset.tsx'
import { Input } from '@/components/input.tsx'
import { Textarea } from '@/components/textarea.tsx'
import { Button } from '@/components/button.tsx'
import { useCreateTenant } from '@/api/queries.ts'
import { useCallback } from 'react'
import { CreateTenantRequest } from '@/api'
import { parse } from 'yaml'
import { useNavigate } from 'react-router-dom'
import {
  defaultPath,
  defaultRepositoryURL,
  defaultTargetRevision,
} from '@/pages/tenants/constants.ts'

interface Data {
  repoURL: string
  path: string
  targetRevision: string
  helmValues: string
}

export default function CreateTenantPage() {
  const { mutate } = useCreateTenant()
  const { register, handleSubmit } = useForm<Data>()
  const nav = useNavigate()
  const onSubmit = useCallback(
    (data: Data) => {
      const req: CreateTenantRequest = {
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
    [mutate, nav],
  )

  return (
    <form className="mx-auto max-w-2xl" onSubmit={handleSubmit(onSubmit)}>
      <Heading>Create a new tenant</Heading>
      <div className="space-y-8">
        <Fieldset className="mt-6">
          <FieldGroup>
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
        <Button type="submit">Create tenant</Button>
      </div>
    </form>
  )
}
