import { NextResponse } from 'next/server'
import { createClient } from '@/lib/supabase/server'

export async function GET() {
  try {
    const supabase = await createClient()
    
    // Check database connection
    const { data, error } = await supabase
      .schema('imp')
      .from('roles')
      .select('count')
      .limit(1)
    
    if (error) {
      return NextResponse.json(
        { 
          status: 'unhealthy',
          error: 'Database connection failed',
          details: error.message 
        },
        { status: 503 }
      )
    }

    return NextResponse.json({
      status: 'healthy',
      timestamp: new Date().toISOString(),
      database: 'connected',
      version: process.env.npm_package_version || 'unknown'
    })
  } catch (error) {
    return NextResponse.json(
      { 
        status: 'unhealthy',
        error: 'Health check failed',
        details: error instanceof Error ? error.message : 'Unknown error'
      },
      { status: 503 }
    )
  }
}
