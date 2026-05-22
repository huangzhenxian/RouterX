import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Plus, Power, PowerOff, Trash2, Activity, AlertCircle } from 'lucide-react';
import {
  listProviders,
  createProvider,
  deleteProvider,
  enableProvider,
  disableProvider,
  testProvider,
} from '@/api/providers';
import type { CreateProviderInput } from '@/types/provider';
import { cn, errorMessage } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

const blankForm: CreateProviderInput = {
  name: '',
  type: 'socks5',
  host: '',
  port: 1080,
  username: '',
  password: '',
  region: '',
  priority: 10,
};

export function Providers() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ['providers'],
    queryFn: listProviders,
    refetchInterval: 15_000,
  });
  const providers = data?.items ?? [];

  const [open, setOpen] = useState(false);
  const [form, setForm] = useState<CreateProviderInput>(blankForm);
  const [testingId, setTestingId] = useState<number | null>(null);

  const createM = useMutation({
    mutationFn: createProvider,
    onSuccess: () => {
      toast.success('已创建');
      setOpen(false);
      setForm(blankForm);
      qc.invalidateQueries({ queryKey: ['providers'] });
    },
    onError: (e) => toast.error(errorMessage(e, '创建失败')),
  });

  const enableM = useMutation({
    mutationFn: enableProvider,
    onSuccess: () => { toast.success('已启用'); qc.invalidateQueries({ queryKey: ['providers'] }); },
    onError: (e) => toast.error(errorMessage(e, '启用失败')),
  });
  const disableM = useMutation({
    mutationFn: disableProvider,
    onSuccess: () => { toast.success('已禁用'); qc.invalidateQueries({ queryKey: ['providers'] }); },
    onError: (e) => toast.error(errorMessage(e, '禁用失败')),
  });
  const deleteM = useMutation({
    mutationFn: deleteProvider,
    onSuccess: () => { toast.success('已删除'); qc.invalidateQueries({ queryKey: ['providers'] }); },
    onError: (e) => toast.error(errorMessage(e, '删除失败')),
  });

  const onTest = async (id: number) => {
    setTestingId(id);
    try {
      const res = await testProvider(id);
      if (res.healthy) {
        toast.success(`可用 · 延迟 ${res.latency_ms} ms`);
      } else {
        toast.error(`不可用：${res.error || '未知错误'}`);
      }
      qc.invalidateQueries({ queryKey: ['providers'] });
    } catch (e) {
      toast.error(errorMessage(e, '测试失败'));
    } finally {
      setTestingId(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">出口代理</h1>
          <p className="text-sm text-muted-foreground mt-1">
            住宅 / 数据中心 SOCKS5/HTTP 代理池，调度器每 2 分钟检查一次
          </p>
        </div>
        <Button onClick={() => setOpen(true)}>
          <Plus className="h-4 w-4" />
          新增代理
        </Button>
      </div>

      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-16">ID</TableHead>
              <TableHead>名称</TableHead>
              <TableHead>协议</TableHead>
              <TableHead>地址</TableHead>
              <TableHead>地区</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>健康</TableHead>
              <TableHead>延迟</TableHead>
              <TableHead>失败</TableHead>
              <TableHead>最近检查</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow><TableCell colSpan={11} className="text-center text-sm text-muted-foreground py-8">加载中…</TableCell></TableRow>
            )}
            {!isLoading && providers.length === 0 && (
              <TableRow><TableCell colSpan={11} className="text-center text-sm text-muted-foreground py-8">还没有代理，点右上角新增一条</TableCell></TableRow>
            )}
            {providers.map((p) => (
              <TableRow key={p.id}>
                <TableCell className="font-mono text-xs">{p.id}</TableCell>
                <TableCell className="font-medium">{p.name}</TableCell>
                <TableCell className="text-xs uppercase">{p.type}</TableCell>
                <TableCell className="text-xs font-mono">{p.host}:{p.port}</TableCell>
                <TableCell className="text-xs">{p.region || '—'}</TableCell>
                <TableCell>
                  {p.status === 1
                    ? <Badge variant="success">启用</Badge>
                    : <Badge variant="secondary">禁用</Badge>}
                </TableCell>
                <TableCell>
                  {p.last_checked_at == null ? (
                    <span className="text-muted-foreground text-xs">未测</span>
                  ) : p.healthy ? (
                    <Badge variant="success">健康</Badge>
                  ) : (
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="inline-flex items-center gap-1 text-xs text-destructive">
                          <AlertCircle className="h-3 w-3" /> 异常
                        </span>
                      </TooltipTrigger>
                      <TooltipContent className="max-w-xs">{p.last_error || '无错误信息'}</TooltipContent>
                    </Tooltip>
                  )}
                </TableCell>
                <TableCell className="tabular-nums text-xs">
                  {p.latency_ms ? `${p.latency_ms} ms` : '—'}
                </TableCell>
                <TableCell className="tabular-nums text-xs">
                  {p.fail_count > 0
                    ? <span className="text-destructive">{p.fail_count}</span>
                    : <span className="text-muted-foreground">0</span>}
                </TableCell>
                <TableCell className="text-xs text-muted-foreground">
                  {p.last_checked_at ? new Date(p.last_checked_at).toLocaleString() : '—'}
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-1">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          size="sm"
                          variant="ghost"
                          disabled={testingId === p.id}
                          onClick={() => onTest(p.id)}
                        >
                          <Activity className={cn('h-4 w-4', testingId === p.id && 'animate-pulse')} />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>立即测试</TooltipContent>
                    </Tooltip>
                    {p.status === 1 ? (
                      <Button size="sm" variant="ghost" onClick={() => disableM.mutate(p.id)}>
                        <PowerOff className="h-4 w-4" />
                      </Button>
                    ) : (
                      <Button size="sm" variant="ghost" onClick={() => enableM.mutate(p.id)}>
                        <Power className="h-4 w-4" />
                      </Button>
                    )}
                    <Button
                      size="sm"
                      variant="ghost"
                      className="text-destructive hover:text-destructive"
                      onClick={() => {
                        if (confirm(`确定删除 ${p.name}？`)) deleteM.mutate(p.id);
                      }}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>新增出口代理</DialogTitle>
            <DialogDescription>SOCKS5 / HTTP / HTTPS 代理，可带用户名密码</DialogDescription>
          </DialogHeader>
          <form
            onSubmit={(e) => { e.preventDefault(); createM.mutate(form); }}
            className="grid grid-cols-2 gap-4"
          >
            <div className="space-y-2 col-span-2">
              <Label htmlFor="name">名称</Label>
              <Input id="name" required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="us-residential-1" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="type">协议</Label>
              <Select
                id="type"
                value={form.type}
                onChange={(e) => setForm({ ...form, type: e.target.value as CreateProviderInput['type'] })}
              >
                <option value="socks5">SOCKS5</option>
                <option value="http">HTTP</option>
                <option value="https">HTTPS</option>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="priority">优先级（数字越小越优先）</Label>
              <Input id="priority" type="number" value={form.priority ?? 10} onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })} />
            </div>
            <div className="space-y-2 col-span-2">
              <Label htmlFor="host">Host</Label>
              <Input id="host" required value={form.host} onChange={(e) => setForm({ ...form, host: e.target.value })} placeholder="proxy.example.com" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="port">Port</Label>
              <Input id="port" type="number" required min={1} max={65535} value={form.port} onChange={(e) => setForm({ ...form, port: Number(e.target.value) })} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="region">地区</Label>
              <Input id="region" value={form.region ?? ''} onChange={(e) => setForm({ ...form, region: e.target.value })} placeholder="us" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="username">用户名（可选）</Label>
              <Input id="username" value={form.username ?? ''} onChange={(e) => setForm({ ...form, username: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">密码（可选）</Label>
              <Input id="password" type="password" value={form.password ?? ''} onChange={(e) => setForm({ ...form, password: e.target.value })} />
            </div>
            <DialogFooter className="col-span-2">
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>取消</Button>
              <Button type="submit" disabled={createM.isPending}>
                {createM.isPending ? '创建中…' : '创建'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
