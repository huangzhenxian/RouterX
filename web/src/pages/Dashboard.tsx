import { useQuery } from '@tanstack/react-query';
import { Users, ShieldCheck, Activity, Server } from 'lucide-react';
import { listUsers } from '@/api/users';
import { listNodes } from '@/api/nodes';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface StatProps {
  label: string;
  value: string | number;
  hint?: string;
  icon: React.ComponentType<{ className?: string }>;
}

function Stat({ label, value, hint, icon: Icon }: StatProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{label}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-semibold tabular-nums">{value}</div>
        {hint && <p className="text-xs text-muted-foreground mt-1">{hint}</p>}
      </CardContent>
    </Card>
  );
}

export function Dashboard() {
  const { data: usersResp } = useQuery({ queryKey: ['users'], queryFn: listUsers, refetchInterval: 10_000 });
  const { data: nodesResp } = useQuery({ queryKey: ['nodes'], queryFn: listNodes, refetchInterval: 10_000 });
  const users = usersResp?.items ?? [];
  const nodes = nodesResp?.items ?? [];
  const enabled = users.filter((u) => u.status === 1).length;
  const onlineNodes = nodes.filter(
    (n) => n.last_seen && Date.now() - new Date(n.last_seen).getTime() < 90_000,
  ).length;
  const totalUsed = users.reduce((s, u) => s + u.used_traffic, 0);
  const totalUsedGB = (totalUsed / 1024 / 1024 / 1024).toFixed(2);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">概览</h1>
        <p className="text-sm text-muted-foreground mt-1">系统全局数据，每 10 秒自动刷新</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Stat label="用户总数"     value={users.length} icon={Users} />
        <Stat label="启用用户"     value={enabled} hint={`/ ${users.length} 总数`} icon={ShieldCheck} />
        <Stat label="累计流量 (GB)" value={totalUsedGB} icon={Activity} />
        <Stat label="在线节点"     value={onlineNodes} hint={`/ ${nodes.length} 总数`} icon={Server} />
      </div>
    </div>
  );
}
