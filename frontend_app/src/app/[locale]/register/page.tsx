import { Metadata } from 'next';
import RegisterForm from '@/components/RegisterForm';

export const metadata: Metadata = {
  robots: {
    index: false,
    follow: true,
  },
};

export default function RegisterPage() {
  return <RegisterForm />;
}
