import { Layout, Menu, Button, Space } from 'antd';
import { DashboardOutlined, TeamOutlined, CloudServerOutlined, LogoutOutlined } from '@ant-design/icons';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';

const { Sider, Header, Content } = Layout;

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: <Link to="/dashboard">概览</Link> },
  { key: '/users',     icon: <TeamOutlined />,      label: <Link to="/users">用户</Link> },
  { key: '/nodes',     icon: <CloudServerOutlined />, label: <Link to="/nodes">节点</Link> },
];

export function AppLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const admin = useAuthStore((s) => s.admin);
  const clear = useAuthStore((s) => s.clear);

  const onLogout = () => {
    clear();
    navigate('/login', { replace: true });
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider theme="dark" breakpoint="lg" collapsible>
        <div style={{ color: '#fff', textAlign: 'center', padding: 16, fontSize: 18, fontWeight: 600 }}>
          RouteX
        </div>
        <Menu theme="dark" mode="inline" selectedKeys={[location.pathname]} items={menuItems} />
      </Sider>
      <Layout>
        <Header style={{
          background: '#fff',
          padding: '0 24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          <span>代理管理后台</span>
          <Space>
            <span style={{ color: '#666' }}>{admin?.username}</span>
            <Button icon={<LogoutOutlined />} size="small" onClick={onLogout}>退出</Button>
          </Space>
        </Header>
        <Content style={{ margin: 24, background: '#fff', padding: 24, borderRadius: 8 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
