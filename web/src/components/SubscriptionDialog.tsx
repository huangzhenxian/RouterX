import { useQuery } from '@tanstack/react-query';
import { QRCodeSVG } from 'qrcode.react';
import { toast } from 'sonner';
import { Copy, Link as LinkIcon } from 'lucide-react';
import { getUserSubscription } from '@/api/subscription';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

interface Props {
  userId: number | null;
  username?: string;
  onClose: () => void;
}

export function SubscriptionDialog({ userId, username, onClose }: Props) {
  const open = userId !== null;
  const { data, isLoading } = useQuery({
    queryKey: ['subscription', userId],
    queryFn: () => getUserSubscription(userId!),
    enabled: open,
  });

  const copy = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    toast.success(`${label} 已复制`);
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>订阅 — {username}</DialogTitle>
          <DialogDescription>
            扫码或复制订阅地址到 V2rayN / Shadowrocket / Clash Verge 等客户端
          </DialogDescription>
        </DialogHeader>

        {isLoading && (
          <p className="text-sm text-muted-foreground py-6 text-center">加载中…</p>
        )}

        {data && (
          <div className="space-y-5">
            {/* 订阅地址 + QR */}
            <div className="grid grid-cols-[auto_1fr] gap-4 items-center">
              <div className="rounded-md border bg-white p-2">
                <QRCodeSVG value={data.subscription_url} size={160} level="M" />
              </div>
              <div className="space-y-2 min-w-0">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <LinkIcon className="h-3 w-3" /> 订阅地址（推荐方式）
                </div>
                <pre className="text-xs font-mono bg-muted rounded-md p-2 break-all whitespace-pre-wrap">
                  {data.subscription_url}
                </pre>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" onClick={() => copy(data.subscription_url, '订阅地址')}>
                    <Copy className="h-3 w-3" />
                    复制订阅地址
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => copy(data.payload.base64, 'base64 订阅')}>
                    复制 base64
                  </Button>
                </div>
              </div>
            </div>

            {/* 每节点单独的 vless URL */}
            <div className="space-y-2">
              <div className="text-xs text-muted-foreground">每个节点的 vless:// 链接（不想用订阅的可单独导入）</div>
              <div className="space-y-2">
                {data.payload.links.map((l) => (
                  <div key={l.node_name} className="rounded-md border p-3 space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="text-sm font-medium">
                        {l.node_name}
                        <span className="text-muted-foreground font-normal text-xs ml-2">{l.host}:{l.port}</span>
                      </div>
                      <Button size="sm" variant="ghost" onClick={() => copy(l.url, `${l.node_name} 链接`)}>
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                    <pre className="text-xs font-mono text-muted-foreground break-all whitespace-pre-wrap">
                      {l.url}
                    </pre>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
