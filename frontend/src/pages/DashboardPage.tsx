import { useState } from 'react';
import { AppShell, Box, Group, Title, ActionIcon, Stack, Button, Text, TextInput, Center } from '@mantine/core';

import { modals } from '@mantine/modals';
import { notifications } from '@mantine/notifications';
import { TbTrash, TbEdit, TbInfoCircle } from 'react-icons/tb';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import { Sidebar } from '../components/Sidebar';
import { SystemMetrics } from '../components/SystemMetrics';
import { TrafficChart } from '../components/TrafficChart';
import { TokensModal } from '../components/TokensModal';

import { getNodes, addNode, updateNode, deleteNode, getNodeTraffic } from '../api/client';

export default function DashboardPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [selectedId, setSelectedId] = useState<number | null>(null);
  
  const { data: nodes = [], isLoading: nodesLoading } = useQuery({
    queryKey: ['nodes'],
    queryFn: getNodes,
  });

  const { data: trafficData = [] } = useQuery({
    queryKey: ['traffic', selectedId],
    queryFn: () => getNodeTraffic(selectedId!),
    enabled: !!selectedId,
  });

  const selectedNode = nodes.find((n) => n.id === selectedId) || null;

  const addMutation = useMutation({
    mutationFn: addNode,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
      modals.closeAll();
      notifications.show({ title: t('common.success'), message: 'Node added successfully', color: 'teal' });
    },
    onError: (err: any) => {
      notifications.show({ title: t('common.error'), message: err.message, color: 'red' });
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: { name: string; url: string } }) => updateNode(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
      modals.closeAll();
      notifications.show({ title: t('common.success'), message: 'Node updated successfully', color: 'teal' });
    },
    onError: (err: any) => {
      notifications.show({ title: t('common.error'), message: err.message, color: 'red' });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteNode,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
      setSelectedId(null);
      modals.closeAll();
      notifications.show({ title: t('common.success'), message: 'Node deleted', color: 'teal' });
    },
  });

  const openNodeModal = (node?: typeof selectedNode) => {
    let name = node?.name || '';
    let url = node?.url || '';

    modals.open({
      title: node ? t('modal.editTitle') : t('modal.addTitle'),
      children: (
        <Stack gap="md">
          <TextInput
            label={t('modal.name')}
            placeholder={t('modal.namePlaceholder')}
            defaultValue={name}
            onChange={(e) => (name = e.currentTarget.value)}
            required
          />
          <TextInput
            label={t('modal.url')}
            placeholder={t('modal.urlPlaceholder')}
            defaultValue={url}
            onChange={(e) => (url = e.currentTarget.value)}
            required
          />
          <Group justify="flex-end" mt="md">
            <Button variant="default" onClick={() => modals.closeAll()}>{t('modal.cancel')}</Button>
            <Button
              color="cyan"
              onClick={() => {
                if (node) {
                  updateMutation.mutate({ id: node.id, data: { name, url } });
                } else {
                  addMutation.mutate({ name, url });
                }
              }}
            >
              {t('modal.save')}
            </Button>
          </Group>
        </Stack>
      ),
    });
  };

  const confirmDelete = (id: number) => {
    modals.openConfirmModal({
      title: t('dashboard.delete'),
      children: <Text size="sm">{t('dashboard.confirmDelete')}</Text>,
      labels: { confirm: t('dashboard.delete'), cancel: t('modal.cancel') },
      confirmProps: { color: 'red' },
      onConfirm: () => deleteMutation.mutate(id),
    });
  };

  const openTokensModal = () => {
    modals.open({
      title: t('tokens.title'),
      size: 'lg',
      children: <TokensModal />,
    });
  };

  return (
    <AppShell
      navbar={{ width: 300, breakpoint: 'sm' }}
      style={{ backgroundColor: 'var(--mantine-color-dark-9)' }}
    >
      <AppShell.Navbar>
        <Sidebar
          nodes={nodes}
          selectedId={selectedId}
          onSelect={setSelectedId}
          onAdd={() => openNodeModal()}
          onManageTokens={openTokensModal}
        />
      </AppShell.Navbar>

      <AppShell.Main>
        <Box p="xl" style={{ maxWidth: 1200, margin: '0 auto' }}>
          {!selectedNode ? (
            <Center h="calc(100vh - 100px)">
              <Stack align="center" gap="sm">
                <TbInfoCircle size={48} color="var(--mantine-color-dark-4)" />
                <Title order={3} c="dimmed">{t('dashboard.selectNode')}</Title>
              </Stack>
            </Center>
          ) : (
            <Stack gap="xl">
              <Group justify="space-between" align="flex-start">
                <Box>
                  <Title order={2} c="white">{selectedNode.name}</Title>
                  <Text c="dimmed" size="sm" mt={4}>{selectedNode.url}</Text>
                </Box>
                <Group>
                  <ActionIcon variant="light" color="cyan" size="lg" onClick={() => openNodeModal(selectedNode)}>
                    <TbEdit size={20} />
                  </ActionIcon>
                  <ActionIcon variant="light" color="red" size="lg" onClick={() => confirmDelete(selectedNode.id)}>
                    <TbTrash size={20} />
                  </ActionIcon>
                </Group>
              </Group>

              <SystemMetrics node={selectedNode} loading={nodesLoading} />

              <TrafficChart data={trafficData} />
            </Stack>
          )}
        </Box>
      </AppShell.Main>
    </AppShell>
  );
}
