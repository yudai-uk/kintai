import { Header } from '@/components/Header';
import { LoginForm } from './LoginForm';

export default async function LoginPage({
  searchParams,
}: {
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  const params = searchParams ? await searchParams : undefined;
  const redirectTo = (params?.redirect as string) || '/employee';
  return (
    <div className="min-h-screen bg-neutral-50 text-[#141414]">
      <Header />
      <main className="px-6 md:px-10 flex justify-center py-5">
        <div className="w-full max-w-[512px] py-5">
          <h1 className="text-[#141414] text-[28px] font-bold leading-tight px-4 text-center pb-3 pt-5">
            Log in to TimeTrack
          </h1>
          <LoginForm redirect={redirectTo} />
        </div>
      </main>
    </div>
  );
}
