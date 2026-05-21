import { Layout, Menu } from 'antd';
import { DashboardOutlined, TeamOutlined, CloudServerOutlined } from '@ant-design/icons';
import { Link, Outlet, useLocation } from 'react-router-dom';

const { Sider, Header, Content } = Layout;

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: <Link to="/dashboard">概览</Link> },
  { key: '/users',     icon: <TeamOutlined />,      label: <Link to="/users">用户</Link> },
  { key: '/nodes',     icon: <CloudServerOutlined />, label: <Link to="/nodes">节点</Link> },
];

export function AppLayout() {
  const location = useLocation();
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider theme="dark" breakpoint="lg" collapsible>
        <div style={{ color: '#fff', textAlign: 'center', padding: 16, fontSize: 18, fontWeight: 600 }}>
          RouteX
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', paddingLeft: 24 }}>代理管理后台</Header>
        <Content style={{ margin: 24, background: '#fff', padding: 24, borderRadius: 8 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
