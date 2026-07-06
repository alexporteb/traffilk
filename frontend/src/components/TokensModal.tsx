import { useState } from 'react';
import { Stack, Group, TextInput, Button, Table, ActionIcon, CopyButton, Text, Alert } from '@mantine/core';
import { TbTrash, TbCopy, TbCheck, TbInfoCircle } from 'react-icons/tb';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { getTokens, createToken, deleteToken } from '../api/client';
import { notifications } from '@mantine/notifications';

export function TokensModal() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [name, setName] = useState('');
  const [newToken, setNewToken] = useState<string | null>(null);

  const { data: tokens, isLoading } = useQuery({
    queryKey: ['tokens'],
    queryFn: getTokens,
  });

  const createMutation = useMutation({
    mutationFn: createToken,
    onSuccess: (data) => {
      setNewToken(data.token);
      setName('');
      queryClient.invalidateQueries({ queryKey: ['tokens'] });
    },
    onError: (err: any) => {
      notifications.show({ title: t('common.error'), message: err.message, color: 'red' });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteToken,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tokens'] });
    },
  });

  return (
    <Stack gap="md">
      {newToken && (
        <Alert icon={<TbInfoCircle />} title={t('tokens.warning')} color="yellow" variant="light">
          <Group align="center" gap="xs" mt="xs">
            <Text ff="monospace" size="sm" style={{ wordBreak: 'break-all' }}>{newToken}</Text>
            <CopyButton value={newToken}>
              {({ copied, copy }) => (
                <Button color={copied ? 'teal' : 'cyan'} onClick={copy} size="xs" variant="light" leftSection={copied ? <TbCheck size={14} /> : <TbCopy size={14} />}>
                  {copied ? t('tokens.copied') : t('tokens.copy')}
                </Button>
              )}
            </CopyButton>
          </Group>
        </Alert>
      )}

      <Group align="flex-end">
        <TextInput
          label={t('tokens.namePlaceholder')}
          value={name}
          onChange={(e) => setName(e.currentTarget.value)}
          style={{ flex: 1 }}
        />
        <Button onClick={() => createMutation.mutate(name)} disabled={!name || createMutation.isPending} color="cyan">
          {t('tokens.create')}
        </Button>
      </Group>

      <Table highlightOnHover>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Name</Table.Th>
            <Table.Th>Created At</Table.Th>
            <Table.Th w={80}></Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {tokens?.map((token) => (
            <Table.Tr key={token.id}>
              <Table.Td>{token.name}</Table.Td>
              <Table.Td>{new Date(token.created_at).toLocaleString()}</Table.Td>
              <Table.Td>
                <ActionIcon color="red" variant="subtle" onClick={() => deleteMutation.mutate(token.id)}>
                  <TbTrash size={16} />
                </ActionIcon>
              </Table.Td>
            </Table.Tr>
          ))}
          {(!tokens || tokens.length === 0) && !isLoading && (
            <Table.Tr>
              <Table.Td colSpan={3} ta="center" c="dimmed">{t('tokens.noTokens')}</Table.Td>
            </Table.Tr>
          )}
        </Table.Tbody>
      </Table>
    </Stack>
  );
}
