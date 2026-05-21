import { Card, Col, Row, Statistic } from 'antd';

export function Dashboard() {
  return (
    <Row gutter={16}>
      <Col span={6}><Card><Statistic title="用户总数" value={0} /></Card></Col>
      <Col span={6}><Card><Statistic title="在线节点" value={0} /></Card></Col>
      <Col span={6}><Card><Statistic title="今日上行 (GB)" value={0} precision={2} /></Card></Col>
      <Col span={6}><Card><Statistic title="今日下行 (GB)" value={0} precision={2} /></Card></Col>
    </Row>
  );
}
