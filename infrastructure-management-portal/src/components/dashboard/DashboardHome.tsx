'use client'

export default function DashboardHome() {
  const quickLinks = [
    { name: 'Servers', description: 'Manage server inventory', icon: '🖥️', href: '/dashboard/servers' },
    { name: 'SSL Certificates', description: 'Track SSL certificates', icon: '🔒', href: '/dashboard/ssl-certificates' },
    { name: 'Applications', description: 'Application catalog', icon: '📱', href: '/dashboard/applications' },
    { name: 'Services', description: 'Services and dependencies', icon: '⚙️', href: '/dashboard/services' },
  ]

  const adminLinks = [
    { name: 'Users', description: 'Manage users and roles', icon: '👥', href: '/dashboard/users' },
    { name: 'Data Models', description: 'Create custom models', icon: '📊', href: '/dashboard/data-models' },
    { name: 'Audit Logs', description: 'View system activity', icon: '📝', href: '/dashboard/audit-logs' },
  ]

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-2">Welcome to Infrastructure Management Portal</h2>
        <p className="text-gray-600">
          Manage your infrastructure data, track resources, and maintain an organized inventory of servers, applications, and services.
        </p>
      </div>

      {/* Core Features */}
      <section>
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Core Features</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {quickLinks.map((link) => (
            <a
              key={link.name}
              href={link.href}
              className="bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
            >
              <div className="text-4xl mb-3">{link.icon}</div>
              <h4 className="text-lg font-semibold text-gray-900 mb-1">{link.name}</h4>
              <p className="text-sm text-gray-600">{link.description}</p>
            </a>
          ))}
        </div>
      </section>

      {/* Admin Section */}
      <section>
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Administration</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {adminLinks.map((link) => (
            <a
              key={link.name}
              href={link.href}
              className="bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
            >
              <div className="text-4xl mb-3">{link.icon}</div>
              <h4 className="text-lg font-semibold text-gray-900 mb-1">{link.name}</h4>
              <p className="text-sm text-gray-600">{link.description}</p>
            </a>
          ))}
        </div>
      </section>

      {/* Getting Started */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-blue-900 mb-2">Getting Started</h3>
        <ul className="space-y-2 text-sm text-blue-800">
          <li>• <strong>Servers:</strong> Add your server inventory with hostname, IP, and specifications</li>
          <li>• <strong>SSL Certificates:</strong> Track certificate expiration dates and SANs</li>
          <li>• <strong>Applications:</strong> Catalog your applications with versions and documentation links</li>
          <li>• <strong>Services:</strong> Map services to servers and track dependencies</li>
          <li>• <strong>Custom Models:</strong> Create additional data models as needed via the Data Models page</li>
        </ul>
      </div>

      {/* Status Info */}
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-3">System Status</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 text-center">
          <div>
            <div className="text-3xl font-bold text-blue-600">0</div>
            <div className="text-sm text-gray-600">Servers</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-green-600">0</div>
            <div className="text-sm text-gray-600">SSL Certificates</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-purple-600">0</div>
            <div className="text-sm text-gray-600">Applications</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-orange-600">0</div>
            <div className="text-sm text-gray-600">Services</div>
          </div>
        </div>
      </div>
    </div>
  )
}
