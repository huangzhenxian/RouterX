import { useState } from 'react';
import { Table, Button, Space, Modal, Form, Input, InputNumber, Popconfirm, Tag, message, Tooltip } from 'antd';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { listUsers, createUser, deleteUser, enableUser, disableUser } from '@/api/users';
import type { User, CreateUserInput } from '@/types/user';

function formatBytes(n: number): string {
  if (!n) return '0';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
  return `${v.toFixed(2)} ${units[i]}`;
}

export function Users() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: listUsers,
    refetchInterval: 10_000,
  });

  const [open, setOpen] = useState(false);
  const [form] = Form.useForm<CreateUserInput>();

  const createM = useMutation({
    mutationFn: createUser,
    onSuccess: () => {
      message.success('创建成功');
      setOpen(false);
      form.resetFields();
      qc.invalidateQueries({ queryKey: ['users'] });
    },
    onError: (e: { response?: { data?: { error?: string } } }) => {
      message.error(e?.response?.data?.error ?? '创建失败');
    },
  });

  const mkMutation = (fn: (id: number) => Promise<void>, label: string) =>
    useMutation({
      mutationFn: fn,
      onSuccess: () => {
        message.success(`${label}成功`);
        qc.invalidateQueries({ queryKey: ['users'] });
      },
      onError: (e: { response?: { data?: { error?: string } } }) => {
        message.error(e?.response?.data?.error ?? `${label}失败`);
      },
    });
  const deleteM  = mkMutation(deleteUser,  '删除');
  const enableM  = mkMutation(enableUser,  '启用');
  const disableM = mkMutation(disableUser, '禁用');

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="middle">
      <Space>
        <Button type="primary" onClick={() => setOpen(true)}>新建用户</Button>
      </Space>

      <Table<User>
        rowKey="id"
        loading={isLoading}
        dataSource={data?.items ?? []}
        pagination={{ pageSize: 20 }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 70 },
          { title: '用户名', dataIndex: 'username' },
          {
            title: 'UUID',
            dataIndex: 'uuid',
            render: (v: string) => (
              <Tooltip title={v}>
                <code style={{ fontSize: 12 }}>{v.slice(0, 8)}…</code>
              </Tooltip>
            ),
          },
          {
            title: '状态',
            dataIndex: 'status',
            width: 80,
            render: (s: number) => s === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>,
          },
          {
            title: '已用 / 配额',
            render: (_, r) => `${formatBytes(r.used_traffic)} / ${r.traffic_limit ? formatBytes(r.traffic_limit) : '∞'}`,
          },
          {
            title: '到期时间',
            dataIndex: 'expire_time',
            render: (v: string) => v && !v.startsWith('0001') ? new Date(v).toLocaleString() : '—',
          },
          {
            title: '操作',
            width: 220,
            render: (_, r) => (
              <Space size="small">
                {r.status === 1
                  ? <Button size="small" onClick={() => disableM.mutate(r.id)}>禁用</Button>
                  : <Button size="small" type="primary" onClick={() => enableM.mutate(r.id)}>启用</Button>}
                <Popconfirm title="确定删除？" onConfirm={() => deleteM.mutate(r.id)}>
                  <Button size="small" danger>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal
        title="新建用户"
        open={open}
        onCancel={() => setOpen(false)}
        onOk={() => form.submit()}
        confirmLoading={createM.isPending}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(v) => createM.mutate(v)}
        >
          <Form.Item label="用户名" name="username" rules={[{ required: true }]}>
            <Input placeholder="alice" />
          </Form.Item>
          <Form.Item
            label="流量配额（字节，0 = 不限）"
            name="traffic_limit"
            tooltip="100GB ≈ 107374182400"
          >
            <InputNumber min={0} step={1024 * 1024 * 1024} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
