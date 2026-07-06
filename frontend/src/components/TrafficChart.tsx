import { useMemo } from 'react';
import { Paper, Title, Box, Text, useMantineTheme } from '@mantine/core';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { useTranslation } from 'react-i18next';
import type { DailyTraffic } from '../api/client';
import { formatBytes } from '../utils/format';

interface TrafficChartProps {
  data: DailyTraffic[];
}

export function TrafficChart({ data }: TrafficChartProps) {
  const { t } = useTranslation();
  const theme = useMantineTheme();

  const chartData = useMemo(() => {
    return data.map((d) => ({
      date: new Date(d.date).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
      rx: d.rx_bytes,
      tx: d.tx_bytes,
    }));
  }, [data]);

  const customTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <Paper p="sm" shadow="md" style={{ backgroundColor: 'var(--mantine-color-dark-7)', border: '1px solid var(--mantine-color-dark-5)' }}>
          <Text size="sm" fw={500} mb={5}>{label}</Text>
          {payload.map((p: any) => (
            <Text key={p.name} size="sm" c={p.color}>
              {p.name === 'rx' ? t('dashboard.download') : t('dashboard.upload')}: {formatBytes(p.value)}
            </Text>
          ))}
        </Paper>
      );
    }
    return null;
  };

  if (!data || data.length === 0) {
    return (
      <Paper p="md" radius="md" style={{ backgroundColor: 'var(--mantine-color-dark-8)', border: '1px solid var(--mantine-color-dark-6)' }}>
        <Title order={4} mb="lg" c="white">{t('dashboard.trafficHistory')}</Title>
        <Box h={300} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Text c="dimmed">No data</Text>
        </Box>
      </Paper>
    );
  }

  return (
    <Paper p="md" radius="md" style={{ backgroundColor: 'var(--mantine-color-dark-8)', border: '1px solid var(--mantine-color-dark-6)' }}>
      <Title order={4} mb="lg" c="white">{t('dashboard.trafficHistory')}</Title>
      <Box h={300}>
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id="colorRx" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor={theme.colors.cyan[6]} stopOpacity={0.3}/>
                <stop offset="95%" stopColor={theme.colors.cyan[6]} stopOpacity={0}/>
              </linearGradient>
              <linearGradient id="colorTx" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor={theme.colors.violet[6]} stopOpacity={0.3}/>
                <stop offset="95%" stopColor={theme.colors.violet[6]} stopOpacity={0}/>
              </linearGradient>
            </defs>
            <XAxis 
              dataKey="date" 
              stroke="var(--mantine-color-dark-3)" 
              fontSize={12} 
              tickLine={false} 
              axisLine={false} 
              tickMargin={10}
            />
            <YAxis 
              tickFormatter={(val) => formatBytes(val, 0)} 
              stroke="var(--mantine-color-dark-3)" 
              fontSize={12} 
              tickLine={false} 
              axisLine={false}
              width={80}
              tickMargin={10}
            />
            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--mantine-color-dark-6)" />
            <Tooltip content={customTooltip} />
            <Area type="monotone" dataKey="tx" stroke={theme.colors.violet[6]} fillOpacity={1} fill="url(#colorTx)" />
            <Area type="monotone" dataKey="rx" stroke={theme.colors.cyan[6]} fillOpacity={1} fill="url(#colorRx)" />
          </AreaChart>
        </ResponsiveContainer>
      </Box>
    </Paper>
  );
}
