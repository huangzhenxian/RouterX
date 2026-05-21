import { Table, Button, Space } from 'antd';

export function Nodes() {
  return (
    <Space direction="vertical" style={{ width: '100%' }} size="middle">
      <Space>
        <Button type="primary">新增节点</Button>
      </Space>
      <Table
        rowKey="id"
        dataSource={[]}
        columns={[
          { title: 'ID', dataIndex: 'id' },
          { title: '名称', dataIndex: 'name' },
          { title: 'IP', dataIndex: 'ip' },
          { title: '地区', dataIndex: 'region' },
          { title: '状态', dataIndex: 'status' },
          { title: 'CPU', dataIndex: 'cpu' },
          { title: '内存', dataIndex: 'memory' },
        ]}
      />
    </Space>
  );
}
