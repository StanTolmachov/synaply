import { Metadata } from 'next';
import LoginForm from '@/components/LoginForm';

export const metadata: Metadata = {
  robots: {
    index: false,
    follow: true,
  },
};

export default function LoginPage() {
  return <LoginForm />;
}
