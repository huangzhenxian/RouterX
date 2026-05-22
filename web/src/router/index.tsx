import { createBrowserRouter, Navigate } from 'react-router-dom';
import { AppLayout } from '@/components/AppLayout';
import { RequireAuth } from '@/components/RequireAuth';
import { Dashboard } from '@/pages/Dashboard';
import { Users } from '@/pages/Users';
import { Nodes } from '@/pages/Nodes';
import { Providers } from '@/pages/Providers';
import { Login } from '@/pages/Login';

export const router = createBrowserRouter([
  { path: '/login', element: <Login /> },
  {
    path: '/',
    element: (
      <RequireAuth>
        <AppLayout />
      </RequireAuth>
    ),
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: <Dashboard /> },
      { path: 'users', element: <Users /> },
      { path: 'nodes', element: <Nodes /> },
      { path: 'providers', element: <Providers /> },
    ],
  },
]);
