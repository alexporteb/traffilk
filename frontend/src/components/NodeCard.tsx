import { Paper, Group, Text, Badge, Stack, RingProgress } from '@mantine/core';

import type { Node } from '../api/client';
import { formatSpeed } from '../utils/format';

interface NodeCardProps {
  node: Node;
  active: boolean;
  onClick: () => void;
}

export function NodeCard({ node, active, onClick }: NodeCardProps) {
  const isUp = node.status === 'up';

  const getTrafficPercent = () => {
    if (!node.isTrafficTrackingActive || node.trafficLimitBytes === 0) return 0;
    return (node.trafficUsedBytes / node.trafficLimitBytes) * 100;
  };

  const trafficPercent = getTrafficPercent();

  return (
    <Paper
      p="sm"
      radius="md"
      onClick={onClick}
      style={{
        cursor: 'pointer',
        backgroundColor: active ? 'var(--mantine-color-dark-6)' : 'var(--mantine-color-dark-8)',
        border: `1px solid ${active ? 'var(--mantine-color-cyan-6)' : 'var(--mantine-color-dark-6)'}`,
        transition: 'all 0.2s ease',
      }}
    >
      <Group justify="space-between" align="center" wrap="nowrap">
        <Stack gap={4} style={{ flex: 1, minWidth: 0 }}>
          <Group gap="xs" wrap="nowrap">
            <Text fw={600} size="sm" truncate c="white">
              {node.name}
            </Text>
            <Badge size="xs" variant="light" color={isUp ? 'cyan' : 'red'}>
              {isUp ? 'Online' : 'Offline'}
            </Badge>
          </Group>
          
          <Group gap="md">
            <Text size="xs" c="dimmed">
              ↓ {formatSpeed(node.rxBytesPerSec)}
            </Text>
            <Text size="xs" c="dimmed">
              ↑ {formatSpeed(node.txBytesPerSec)}
            </Text>
          </Group>
        </Stack>

        {node.isTrafficTrackingActive && (
          <RingProgress
            size={40}
            thickness={4}
            roundCaps
            sections={[{ value: trafficPercent, color: trafficPercent > 90 ? 'red' : 'cyan' }]}
            label={
              <Text style={{ fontSize: 9 }} ta="center" fw={700}>
                {trafficPercent.toFixed(0)}%
              </Text>
            }
          />
        )}
      </Group>
    </Paper>
  );
}
