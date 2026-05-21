import { useState } from 'react';
import { Table, Button, Space, Modal, Form, Input, Popconfirm, Tag, message, Alert } from 'antd';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { listNodes, createNode, deleteNode } from '@/api/nodes';
import type { Node, CreateNodeInput } from '@/types/node';

function isOnline(lastSeen: string | null): boolean {
  if (!lastSeen) return false;
  return Date.now() - new Date(lastSeen).getTime() < 90_000; // 90s
}

export function Nodes() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ['nodes'],
    queryFn: listNodes,
    refetchInterval: 5_000,
  });

  const [open, setOpen] = useState(false);
  const [form] = Form.useForm<CreateNodeInput>();
  const [newToken, setNewToken] = useState<string | null>(null);

  const createM = useMutation({
    mutationFn: createNode,
    onSuccess: (res) => {
      message.success('节点已创建');
      setOpen(false);
      form.resetFields();
      setNewToken(res.auth_token);
      qc.invalidateQueries({ queryKey: ['nodes'] });
    },
    onError: (e: { response?: { data?: { error?: string } } }) => {
      message.error(e?.response?.data?.error ?? '创建失败');
    },
  });

  const deleteM = useMutation({
    mutationFn: deleteNode,
    onSuccess: () => {
      message.success('删除成功');
      qc.invalidateQueries({ queryKey: ['nodes'] });
    },
  });

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="middle">
      <Space>
        <Button type="primary" onClick={() => setOpen(true)}>新增节点</Button>
      </Space>

      <Table<Node>
        rowKey="id"
        loading={isLoading}
        dataSource={data?.items ?? []}
        pagination={{ pageSize: 20 }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 70 },
          { title: '名称', dataIndex: 'name' },
          { title: 'IP', dataIndex: 'ip' },
          { title: '地区', dataIndex: 'region' },
          {
            title: '在线',
            render: (_, r) => isOnline(r.last_seen)
              ? <Tag color="green">在线</Tag>
              : <Tag color="default">离线</Tag>,
          },
          { title: 'CPU%', dataIndex: 'cpu', render: (v: number) => v?.toFixed(1) ?? '—' },
          { title: '内存%', dataIndex: 'memory', render: (v: number) => v?.toFixed(1) ?? '—' },
          { title: '版本', dataIndex: 'version' },
          {
            title: '最近心跳',
            dataIndex: 'last_seen',
            render: (v: string | null) => v ? new Date(v).toLocaleString() : '—',
          },
          {
            title: '操作',
            width: 100,
            render: (_, r) => (
              <Popconfirm title="确定删除？" onConfirm={() => deleteM.mutate(r.id)}>
                <Button size="small" danger>删除</Button>
              </Popconfirm>
            ),
          },
        ]}
      />

      <Modal
        title="新增节点"
        open={open}
        onCancel={() => setOpen(false)}
        onOk={() => form.submit()}
        confirmLoading={createM.isPending}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={(v) => createM.mutate(v)}>
          <Form.Item label="名称" name="name" rules={[{ required: true }]}>
            <Input placeholder="hk-1" />
          </Form.Item>
          <Form.Item label="IP" name="ip"><Input placeholder="1.2.3.4" /></Form.Item>
          <Form.Item label="地区" name="region"><Input placeholder="hk" /></Form.Item>
        </Form>
      </Modal>

      <Modal
        title="节点 token（仅此一次显示，请妥善保存）"
        open={!!newToken}
        onCancel={() => setNewToken(null)}
        onOk={() => setNewToken(null)}
        cancelButtonProps={{ style: { display: 'none' } }}
      >
        <Alert
          type="warning"
          message="把下面这串配给 agent 的 X-Node-Token 头部使用。关闭后无法再次查看。"
          style={{ marginBottom: 12 }}
        />
        <pre style={{ background: '#f5f5f5', padding: 12, borderRadius: 4, wordBreak: 'break-all' }}>
          {newToken}
        </pre>
      </Modal>
    </Space>
  );
}
