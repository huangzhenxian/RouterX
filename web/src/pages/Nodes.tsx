import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Plus, Trash2, Copy, AlertTriangle } from 'lucide-react';
import { listNodes, createNode, deleteNode } from '@/api/nodes';
import type { CreateNodeInput } from '@/types/node';
import { errorMessage } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

function isOnline(lastSeen: string | null): boolean {
  if (!lastSeen) return false;
  return Date.now() - new Date(lastSeen).getTime() < 90_000;
}

export function Nodes() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ['nodes'],
    queryFn: listNodes,
    refetchInterval: 5_000,
  });
  const nodes = data?.items ?? [];

  const [open, setOpen] = useState(false);
  const [form, setForm] = useState<CreateNodeInput>({ name: '', ip: '', region: '' });
  const [revealedToken, setRevealedToken] = useState<string | null>(null);

  const createM = useMutation({
    mutationFn: createNode,
    onSuccess: (res) => {
      toast.success('节点已创建');
      setOpen(false);
      setForm({ name: '', ip: '', region: '' });
      setRevealedToken(res.auth_token);
      qc.invalidateQueries({ queryKey: ['nodes'] });
    },
    onError: (e) => toast.error(errorMessage(e, '创建失败')),
  });

  const deleteM = useMutation({
    mutationFn: deleteNode,
    onSuccess: () => { toast.success('已删除'); qc.invalidateQueries({ queryKey: ['nodes'] }); },
    onError: (e) => toast.error(errorMessage(e, '删除失败')),
  });

  const copyToken = () => {
    if (revealedToken) {
      navigator.clipboard.writeText(revealedToken);
      toast.success('Token 已复制到剪贴板');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">节点</h1>
          <p className="text-sm text-muted-foreground mt-1">代理节点列表与心跳状态，每 5 秒刷新</p>
        </div>
        <Button onClick={() => setOpen(true)}>
          <Plus className="h-4 w-4" />
          新增节点
        </Button>
      </div>

      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-16">ID</TableHead>
              <TableHead>名称</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>地区</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>CPU%</TableHead>
              <TableHead>内存%</TableHead>
              <TableHead>版本</TableHead>
              <TableHead>最近心跳</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow><TableCell colSpan={10} className="text-center text-sm text-muted-foreground py-8">加载中…</TableCell></TableRow>
            )}
            {!isLoading && nodes.length === 0 && (
              <TableRow><TableCell colSpan={10} className="text-center text-sm text-muted-foreground py-8">暂无节点</TableCell></TableRow>
            )}
            {nodes.map((n) => (
              <TableRow key={n.id}>
                <TableCell className="font-mono text-xs">{n.id}</TableCell>
                <TableCell className="font-medium">{n.name}</TableCell>
                <TableCell className="text-xs">{n.ip || '—'}</TableCell>
                <TableCell className="text-xs">{n.region || '—'}</TableCell>
                <TableCell>
                  {isOnline(n.last_seen)
                    ? <Badge variant="success">在线</Badge>
                    : <Badge variant="secondary">离线</Badge>}
                </TableCell>
                <TableCell className="tabular-nums text-xs">{n.cpu?.toFixed(1) ?? '—'}</TableCell>
                <TableCell className="tabular-nums text-xs">{n.memory?.toFixed(1) ?? '—'}</TableCell>
                <TableCell className="text-xs text-muted-foreground">{n.version || '—'}</TableCell>
                <TableCell className="text-xs text-muted-foreground">
                  {n.last_seen ? new Date(n.last_seen).toLocaleString() : '—'}
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    size="sm"
                    variant="ghost"
                    className="text-destructive hover:text-destructive"
                    onClick={() => {
                      if (confirm(`确定删除 ${n.name}？`)) deleteM.mutate(n.id);
                    }}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>

      {/* 新增节点 */}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>新增节点</DialogTitle>
            <DialogDescription>创建后会生成一个 token 给 agent 鉴权（仅显示一次）</DialogDescription>
          </DialogHeader>
          <form
            onSubmit={(e) => { e.preventDefault(); createM.mutate(form); }}
            className="space-y-4"
          >
            <div className="space-y-2">
              <Label htmlFor="name">名称</Label>
              <Input id="name" required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="hk-1" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ip">IP</Label>
              <Input id="ip" value={form.ip ?? ''} onChange={(e) => setForm({ ...form, ip: e.target.value })} placeholder="1.2.3.4" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="region">地区</Label>
              <Input id="region" value={form.region ?? ''} onChange={(e) => setForm({ ...form, region: e.target.value })} placeholder="hk" />
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

      {/* Token 显示一次 */}
      <Dialog open={!!revealedToken} onOpenChange={(o) => !o && setRevealedToken(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>节点 token</DialogTitle>
            <DialogDescription>把它配给 agent 的 X-Node-Token 头部</DialogDescription>
          </DialogHeader>
          <Alert variant="warning">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>仅此一次显示</AlertTitle>
            <AlertDescription>关闭对话框后无法再次查看。请妥善保存。</AlertDescription>
          </Alert>
          <pre className="rounded-md bg-muted p-3 text-xs font-mono break-all whitespace-pre-wrap">
            {revealedToken}
          </pre>
          <DialogFooter>
            <Button variant="outline" onClick={copyToken}>
              <Copy className="h-4 w-4" />
              复制
            </Button>
            <Button onClick={() => setRevealedToken(null)}>关闭</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
