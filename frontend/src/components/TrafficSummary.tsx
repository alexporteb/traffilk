import React, { useMemo } from 'react';
import { SimpleGrid, Paper, Text, Group } from '@mantine/core';
import { TbArrowDownRight, TbArrowUpRight, TbServer } from 'react-icons/tb';
import { useTranslation } from 'react-i18next';
import { formatBytes } from '../utils/format';

interface DailyTraffic {
  date: string;
  rx_bytes: number;
  tx_bytes: number;
}

interface TrafficSummaryProps {
  data: DailyTraffic[];
}

export const TrafficSummary: React.FC<TrafficSummaryProps> = ({ data }) => {
  const { t } = useTranslation();

  const { today, month, allTime } = useMemo(() => {
    const now = new Date();
    // format to match our data: YYYY-MM-DD
    const todayStr = now.toISOString().split('T')[0];
    const monthStr = todayStr.substring(0, 7); // YYYY-MM

    const res = {
      today: { rx: 0, tx: 0, total: 0 },
      month: { rx: 0, tx: 0, total: 0 },
      allTime: { rx: 0, tx: 0, total: 0 },
    };

    data.forEach((item) => {
      // all time
      res.allTime.rx += item.rx_bytes;
      res.allTime.tx += item.tx_bytes;
      res.allTime.total += (item.rx_bytes + item.tx_bytes);

      // month
      if (item.date.startsWith(monthStr)) {
        res.month.rx += item.rx_bytes;
        res.month.tx += item.tx_bytes;
        res.month.total += (item.rx_bytes + item.tx_bytes);
      }

      // today
      if (item.date === todayStr) {
        res.today.rx += item.rx_bytes;
        res.today.tx += item.tx_bytes;
        res.today.total += (item.rx_bytes + item.tx_bytes);
      }
    });

    return res;
  }, [data]);

  const cards = [
    { label: t('dashboard.today'), stats: today },
    { label: t('dashboard.month'), stats: month },
    { label: t('dashboard.allTime'), stats: allTime },
  ];

  return (
    <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="md" mb="md">
      {cards.map((card, idx) => (
        <Paper key={idx} p="md" radius="md" bg="dark.8" withBorder>
          <Group justify="space-between" mb="xs">
            <Text size="sm" c="dimmed" fw={500}>
              {card.label}
            </Text>
            <TbServer size={20} opacity={0.5} />
          </Group>

          <Text size="xl" fw={700} mb="sm">
            {formatBytes(card.stats.total)}
          </Text>

          <Group grow gap="xs">
            <Group gap="xs" wrap="nowrap">
              <TbArrowDownRight size={16} color="var(--mantine-color-green-5)" />
              <Text size="xs" c="dimmed">
                {formatBytes(card.stats.rx)} {t('dashboard.rx')}
              </Text>
            </Group>
            <Group gap="xs" wrap="nowrap">
              <TbArrowUpRight size={16} color="var(--mantine-color-blue-5)" />
              <Text size="xs" c="dimmed">
                {formatBytes(card.stats.tx)} {t('dashboard.tx')}
              </Text>
            </Group>
          </Group>
        </Paper>
      ))}
    </SimpleGrid>
  );
};
