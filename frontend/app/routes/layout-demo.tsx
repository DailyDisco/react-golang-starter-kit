import { createFileRoute, Outlet } from '@tanstack/react-router';
import CustomDemoLayout from '../layouts/CustomDemoLayout';

export const Route = createFileRoute('/layout-demo')({
  component: LayoutDemo,
});

function LayoutDemo() {
  return <CustomDemoLayout />;
}
