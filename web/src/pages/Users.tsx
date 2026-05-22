import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Plus, Power, PowerOff, Trash2, QrCode } from 'lucide-react';
import { listUsers, createUser, deleteUser, enableUser, disableUser } from '@/api/users';
import type { CreateUserInput } from '@/types/user';
import { cn, errorMessage, formatBytes } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
import { SubscriptionDialog } from '@/components/SubscriptionDialog';

export function Users() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: listUsers,
    refetchInterval: 10_000,
  });
  const users = data?.items ?? [];

  const [open, setOpen] = useState(false);
  const [form, setForm] = useState<CreateUserInput>({ username: '', traffic_limit: 0 });
  const [subFor, setSubFor] = useState<{ id: number; username: string } | null>(null);

  const createM = useMutation({
    mutationFn: createUser,
    onSuccess: () => {
      toast.success('用户已创建');
      setOpen(false);
      setForm({ username: '', traffic_limit: 0 });
      qc.invalidateQueries({ queryKey: ['users'] });
    },
    onError: (e) => toast.error(errorMessage(e, '创建失败')),
  });

  const enableM = useMutation({
    mutationFn: enableUser,
    onSuccess: () => { toast.success('已启用'); qc.invalidateQueries({ queryKey: ['users'] }); },
    onError: (e) => toast.error(errorMessage(e, '启用失败')),
  });
  const disableM = useMutation({
    mutationFn: disableUser,
    onSuccess: () => { toast.success('已禁用'); qc.invalidateQueries({ queryKey: ['users'] }); },
    onError: (e) => toast.error(errorMessage(e, '禁用失败')),
  });
  const deleteM = useMutation({
    mutationFn: deleteUser,
    onSuccess: () => { toast.success('已删除'); qc.invalidateQueries({ queryKey: ['users'] }); },
    onError: (e) => toast.error(errorMessage(e, '删除失败')),
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">用户</h1>
          <p className="text-sm text-muted-foreground mt-1">代理用户管理，新建后自动同步到 Xray 入站</p>
        </div>
        <Button onClick={() => setOpen(true)}>
          <Plus className="h-4 w-4" />
          新建用户
        </Button>
      </div>

      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-16">ID</TableHead>
              <TableHead>用户名</TableHead>
              <TableHead>UUID</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>已用 / 配额</TableHead>
              <TableHead>到期</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-sm text-muted-foreground py-8">加载中…</TableCell>
              </TableRow>
            )}
            {!isLoading && users.length === 0 && (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-sm text-muted-foreground py-8">暂无用户</TableCell>
              </TableRow>
            )}
            {users.map((u) => (
              <TableRow key={u.id}>
                <TableCell className="font-mono text-xs">{u.id}</TableCell>
                <TableCell className="font-medium">{u.username}</TableCell>
                <TableCell>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <code className="text-xs font-mono text-muted-foreground">{u.uuid.slice(0, 8)}…</code>
                    </TooltipTrigger>
                    <TooltipContent><code>{u.uuid}</code></TooltipContent>
                  </Tooltip>
                </TableCell>
                <TableCell>
                  {u.status === 1
                    ? <Badge variant="success">启用</Badge>
                    : <Badge variant="secondary">禁用</Badge>}
                </TableCell>
                <TableCell className="tabular-nums text-xs">
                  {formatBytes(u.used_traffic)} / {u.traffic_limit ? formatBytes(u.traffic_limit) : '∞'}
                </TableCell>
                <TableCell className="text-xs text-muted-foreground">
                  {u.expire_time && !u.expire_time.startsWith('0001') ? new Date(u.expire_time).toLocaleString() : '—'}
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-1">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button size="sm" variant="ghost" onClick={() => setSubFor({ id: u.id, username: u.username })}>
                          <QrCode className="h-4 w-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>订阅 / 二维码</TooltipContent>
                    </Tooltip>
                    {u.status === 1 ? (
                      <Button size="sm" variant="ghost" onClick={() => disableM.mutate(u.id)}>
                        <PowerOff className="h-4 w-4" />
                      </Button>
                    ) : (
                      <Button size="sm" variant="ghost" onClick={() => enableM.mutate(u.id)}>
                        <Power className="h-4 w-4" />
                      </Button>
                    )}
                    <Button
                      size="sm"
                      variant="ghost"
                      className={cn('text-destructive hover:text-destructive')}
                      onClick={() => {
                        if (confirm(`确定删除 ${u.username}？`)) deleteM.mutate(u.id);
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
        <DialogContent>
          <DialogHeader>
            <DialogTitle>新建用户</DialogTitle>
            <DialogDescription>创建后会自动在 Xray 入站里添加这个用户</DialogDescription>
          </DialogHeader>
          <form
            onSubmit={(e) => { e.preventDefault(); createM.mutate(form); }}
            className="space-y-4"
          >
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <Input
                id="username"
                value={form.username}
                onChange={(e) => setForm({ ...form, username: e.target.value })}
                placeholder="alice"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="limit">流量配额（字节，0 = 不限）</Label>
              <Input
                id="limit"
                type="number"
                min={0}
                value={form.traffic_limit ?? 0}
                onChange={(e) => setForm({ ...form, traffic_limit: Number(e.target.value) })}
              />
              <p className="text-xs text-muted-foreground">100 GB ≈ 107374182400</p>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>取消</Button>
              <Button type="submit" disabled={createM.isPending}>
                {createM.isPending ? '创建中…' : '创建'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <SubscriptionDialog
        userId={subFor?.id ?? null}
        username={subFor?.username}
        onClose={() => setSubFor(null)}
      />
    </div>
  );
}
