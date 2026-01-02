import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import DashboardLayout from '@/components/dashboard/DashboardLayout'
import DashboardHome from '@/components/dashboard/DashboardHome'

export default async function DashboardPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect('/auth/login')
  }

  // Get user profile with role
  const { data: profile } = await supabase
    .schema('imp')
    .from('user_profiles')
    .select('*, role:roles(*)')
    .eq('id', user.id)
    .single()

  return (
    <DashboardLayout user={user} profile={profile}>
      <DashboardHome />
    </DashboardLayout>
  )
}
