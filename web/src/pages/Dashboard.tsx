import { Card, Col, Row, Statistic } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { listUsers } from '@/api/users';

export function Dashboard() {
  const { data } = useQuery({ queryKey: ['users'], queryFn: listUsers, refetchInterval: 10_000 });
  const users = data?.items ?? [];
  const enabled = users.filter((u) => u.status === 1).length;
  const totalUsed = users.reduce((s, u) => s + u.used_traffic, 0);
  const totalUsedGB = totalUsed / 1024 / 1024 / 1024;

  return (
    <Row gutter={16}>
      <Col span={6}><Card><Statistic title="用户总数" value={users.length} /></Card></Col>
      <Col span={6}><Card><Statistic title="启用用户" value={enabled} /></Card></Col>
      <Col span={6}><Card><Statistic title="累计已用流量 (GB)" value={totalUsedGB} precision={2} /></Card></Col>
      <Col span={6}><Card><Statistic title="在线节点" value={0} /></Card></Col>
    </Row>
  );
}
