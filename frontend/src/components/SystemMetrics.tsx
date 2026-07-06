import { SimpleGrid, Progress, Text, Stack } from '@mantine/core';
import { useTranslation } from 'react-i18next';
import { TbCpu, TbActivity, TbServer, TbClock, TbWifiOff, TbFiles } from 'react-icons/tb';
import { MetricCard } from './MetricCard';
import { formatBytes, formatUptime, formatPercent } from '../utils/format';
import type { Node } from '../api/client';

interface SystemMetricsProps {
  node: Node | null;
  loading: boolean;
}

export function SystemMetrics({ node, loading }: SystemMetricsProps) {
  const { t } = useTranslation();

  const renderMemoryValue = () => {
    if (!node || node.memTotalBytes === 0) return '—';
    const percent = (node.memUsedBytes / node.memTotalBytes) * 100;
    return (
      <Stack gap={4}>
        <Text size="lg" fw={700} c="white" truncate>
          {formatBytes(node.memUsedBytes)} / {formatBytes(node.memTotalBytes)}
        </Text>
        <Progress value={percent} color="cyan" size="sm" mt={2} />
      </Stack>
    );
  };

  const renderCpuValue = () => {
    if (!node) return '—';
    return (
      <Stack gap={4}>
        <Text size="lg" fw={700} c="white" truncate>
          {formatPercent(node.cpuLoadPercent)}
        </Text>
        <Progress value={node.cpuLoadPercent} color="cyan" size="sm" mt={2} />
      </Stack>
    );
  };

  const getLoadAvg = () => {
    if (!node) return '—';
    return `${node.loadAvg1.toFixed(2)} | ${node.loadAvg5.toFixed(2)} | ${node.loadAvg15.toFixed(2)}`;
  };

  return (
    <SimpleGrid cols={{ base: 1, sm: 2, md: 3 }} spacing="md">
      <MetricCard
        title={t('metrics.cpu')}
        value={renderCpuValue()}
        icon={<TbCpu size={20} />}
        loading={loading}
      />
      <MetricCard
        title={t('metrics.memory')}
        value={renderMemoryValue()}
        icon={<TbServer size={20} />}
        loading={loading}
      />
      <MetricCard
        title={t('metrics.loadAvg')}
        value={getLoadAvg()}
        icon={<TbActivity size={20} />}
        loading={loading}
      />
      <MetricCard
        title={t('metrics.uptime')}
        value={node ? formatUptime(node.uptimeSeconds) : '—'}
        icon={<TbClock size={20} />}
        loading={loading}
      />
      <MetricCard
        title={t('metrics.netDrops')}
        value={node ? `${node.netDropsRx + node.netDropsTx}` : '—'}
        icon={<TbWifiOff size={20} />}
        color={node && (node.netDropsRx + node.netDropsTx) > 0 ? 'red' : 'cyan'}
        loading={loading}
      />
      <MetricCard
        title={t('metrics.fileDescriptors')}
        value={node ? node.fileDescriptors.toString() : '—'}
        icon={<TbFiles size={20} />}
        loading={loading}
      />
    </SimpleGrid>
  );
}
