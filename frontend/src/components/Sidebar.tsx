import { useState } from 'react';
import { Stack, TextInput, Button, ScrollArea, Group, Title, ActionIcon, Menu } from '@mantine/core';
import { TbSearch, TbPlus, TbSettings, TbLogout, TbLanguage } from 'react-icons/tb';
import { useTranslation } from 'react-i18next';
import { NodeCard } from './NodeCard';
import type { Node } from '../api/client';
import { logout } from '../api/client';
import { useNavigate } from 'react-router-dom';

interface SidebarProps {
  nodes: Node[];
  selectedId: number | null;
  onSelect: (id: number) => void;
  onAdd: () => void;
  onManageTokens: () => void;
}


export function Sidebar({ nodes, selectedId, onSelect, onAdd, onManageTokens }: SidebarProps) {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [search, setSearch] = useState('');

  const filteredNodes = nodes.filter((n) =>
    n.name.toLowerCase().includes(search.toLowerCase())
  );

  const handleLogout = async () => {
    try {
      await logout();
    } finally {
      navigate('/login');
    }
  };

  return (
    <Stack h="100%" gap="md" p="md" style={{ backgroundColor: 'var(--mantine-color-dark-8)', borderRight: '1px solid var(--mantine-color-dark-6)' }}>
      <Group justify="space-between">
        <Title order={3} c="white">Traffilk</Title>
        <Menu position="bottom-end" shadow="md">
          <Menu.Target>
            <ActionIcon variant="subtle" color="gray">
              <TbSettings size={20} />
            </ActionIcon>
          </Menu.Target>
          <Menu.Dropdown>
            <Menu.Item leftSection={<TbSettings size={14} />} onClick={onManageTokens}>
              {t('tokens.title')}
            </Menu.Item>
            <Menu.Item leftSection={<TbLanguage size={14} />} onClick={() => i18n.changeLanguage(i18n.language === 'en' ? 'ru' : 'en')}>
              {i18n.language === 'en' ? 'Русский' : 'English'}
            </Menu.Item>
            <Menu.Divider />
            <Menu.Item color="red" leftSection={<TbLogout size={14} />} onClick={handleLogout}>
              {t('common.logout')}
            </Menu.Item>
          </Menu.Dropdown>
        </Menu>
      </Group>

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
