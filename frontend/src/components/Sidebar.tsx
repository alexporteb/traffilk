import { useState } from 'react';
import { Stack, TextInput, Button, ScrollArea, Title } from '@mantine/core';
import { TbSearch, TbPlus } from 'react-icons/tb';
import { useTranslation } from 'react-i18next';
import { NodeCard } from './NodeCard';
import type { Node } from '../api/client';

interface SidebarProps {
  nodes: Node[];
  selectedId: number | null;
  onSelect: (id: number) => void;
  onAdd: () => void;
}


export function Sidebar({ nodes, selectedId, onSelect, onAdd }: SidebarProps) {
  const { t } = useTranslation();
  const [search, setSearch] = useState('');

  const filteredNodes = nodes.filter((n) =>
    n.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <Stack h="100%" gap="md" p="md" style={{ backgroundColor: 'var(--mantine-color-dark-8)', borderRight: '1px solid var(--mantine-color-dark-6)' }}>

      <Button leftSection={<TbPlus size={16} />} variant="light" color="cyan" fullWidth onClick={onAdd}>
        {t('sidebar.addNode')}
      </Button>

      <TextInput
        placeholder={t('sidebar.search')}
        leftSection={<TbSearch size={16} />}
        value={search}
        onChange={(e) => setSearch(e.currentTarget.value)}
        styles={{
          input: {
            backgroundColor: 'var(--mantine-color-dark-7)',
            border: '1px solid var(--mantine-color-dark-5)',
          },
        }}
      />

      <ScrollArea style={{ flex: 1 }} offsetScrollbars>
        <Stack gap="xs">
          {filteredNodes.map((node) => (
            <NodeCard
              key={node.id}
              node={node}
              active={node.id === selectedId}
              onClick={() => onSelect(node.id)}
            />
          ))}
          {filteredNodes.length === 0 && (
            <Title order={6} c="dimmed" ta="center" mt="xl">
              {t('sidebar.noNodes')}
            </Title>
          )}
        </Stack>
      </ScrollArea>
    </Stack>
  );
}
