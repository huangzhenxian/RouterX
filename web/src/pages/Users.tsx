import { Table, Button, Space } from 'antd';

export function Users() {
  return (
    <Space direction="vertical" style={{ width: '100%' }} size="middle">
      <Space>
        <Button type="primary">新建用户</Button>
      </Space>
      <Table
        rowKey="id"
        dataSource={[]}
        columns={[
          { title: 'ID', dataIndex: 'id' },
          { title: '用户名', dataIndex: 'username' },
          { title: 'UUID', dataIndex: 'uuid' },
          { title: '状态', dataIndex: 'status' },
          { title: '已用流量', dataIndex: 'used_traffic' },
          { title: '到期时间', dataIndex: 'expire_time' },
        ]}
      />
    </Space>
  );
}
