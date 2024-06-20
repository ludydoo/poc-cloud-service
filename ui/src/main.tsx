import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import App from '@/App.tsx'
import TenantListPage from '@/pages/tenants/list/TenantListPage.tsx'
import CreateTenantPage from '@/pages/tenants/create/CreateTenantPage.tsx'
import TenantDetailsPage from '@/pages/tenants/details/TenantDetailsPage.tsx'
import EditTenantPage from '@/pages/tenants/edit/EditTenantPage.tsx'

const router = createBrowserRouter([
  {
    element: <App />,
    children: [
      {
        index: true,
        element: <div />,
      },
      {
        path: 'tenants/create',
        element: <CreateTenantPage />,
      },
      {
        path: 'tenants',
        element: <TenantListPage />,
      },
      {
        path: 'tenants/:id',
        element: <TenantDetailsPage />,
      },
      {
        path: 'tenants/:id/edit',
        element: <EditTenantPage />,
      },
    ],
  },
])

const queryClient = new QueryClient()

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </React.StrictMode>,
)
