import { Badge } from '@/components/badge.tsx'

export default function TenantStatus({ status }: { status?: string }) {
  if (!status) {
    return <Badge color="zinc">Unknown</Badge>
  }
  if (status === 'Healthy') {
    return <Badge color="green">Healthy</Badge>
  }
  if (status === 'Progressing') {
    return <Badge color="yellow">Progressing</Badge>
  }
  if (status === 'Missing') {
    return <Badge color="amber">Missing</Badge>
  }
  return null
}
