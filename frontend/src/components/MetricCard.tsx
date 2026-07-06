import { Paper, Group, Text, Stack, ThemeIcon, Skeleton } from '@mantine/core';
import type { ReactNode } from 'react';

interface MetricCardProps {
  title: string;
  value: ReactNode;
  icon: ReactNode;
  color?: string;
  loading?: boolean;
}

export function MetricCard({ title, value, icon, color = 'cyan', loading }: MetricCardProps) {
  return (
    <Paper
      p="md"
      radius="md"
      style={{
        backgroundColor: 'var(--mantine-color-dark-8)',
        border: '1px solid var(--mantine-color-dark-6)',
      }}
    >
      <Group justify="space-between" align="flex-start">
        <Stack gap={2}>
          <Text size="sm" c="dimmed" fw={500}>
            {title}
          </Text>
          {loading ? (
            <Skeleton height={28} width={80} mt={4} />
          ) : (
            <Text size="xl" fw={700} c="white">
              {value}
            </Text>
          )}
        </Stack>
        <ThemeIcon size={38} radius="md" variant="light" color={color}>
          {icon}
        </ThemeIcon>
      </Group>
    </Paper>
  );
}
