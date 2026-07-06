import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Center,
  Stack,
  Paper,
  Title,
  TextInput,
  PasswordInput,
  Button,
  Alert,
  Image,
  Select,
  Group,
  Box,
} from '@mantine/core';
import { TbAlertCircle, TbActivity } from 'react-icons/tb';
import { login } from '../api/client';

export default function LoginPage() {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(username, password);
      navigate('/dashboard');
    } catch (err: any) {
      if (err.message?.includes('Too many')) {
        setError(t('login.rateLimited'));
      } else if (err.message?.includes('Invalid')) {
        setError(t('login.error'));
      } else {
        setError(`Error: ${err.message}`);
      }
    } finally {
      setLoading(false);
    }
  };

  const changeLang = (val: string | null) => {
    if (val) {
      i18n.changeLanguage(val);
      localStorage.setItem('traffilk_lang', val);
    }
  };

  return (
    <Box
      style={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #010409 0%, #0d1117 50%, #010409 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <Center w="100%" px="md">
        <Stack w={380} gap="lg">
          <Stack align="center" gap="xs">
            <Box bg="cyan.9" p="md" style={{ borderRadius: '50%' }}>
              <TbActivity size={48} color="white" />
            </Box>
            <Title order={1} c="white" fw={700} style={{ letterSpacing: 1 }}>
              Traffilk
            </Title>
          </Stack>

          <Paper
            p="xl"
            radius="lg"
            style={{
              backgroundColor: 'var(--mantine-color-dark-7)',
              border: '1px solid var(--mantine-color-dark-5)',
            }}
          >
            <Group justify="space-between" mb="md">
              <Title order={3} c="white">
                {t('login.title')}
              </Title>
              <Select
                data={[
                  { value: 'en', label: 'EN' },
                  { value: 'ru', label: 'RU' },
                ]}
                value={i18n.language}
                onChange={changeLang}
                w={70}
                size="xs"
                styles={{
                  input: {
                    backgroundColor: 'var(--mantine-color-dark-8)',
                    border: '1px solid var(--mantine-color-dark-5)',
                  },
                }}
              />
            </Group>

            {error && (
              <Alert
                icon={<TbAlertCircle />}
                color="red"
                variant="light"
                mb="md"
                radius="md"
              >
                {error}
              </Alert>
            )}

            <form onSubmit={handleSubmit}>
              <Stack gap="md">
                <TextInput
                  label={t('login.username')}
                  value={username}
                  onChange={(e) => setUsername(e.currentTarget.value)}
                  required
                  styles={{
                    input: {
                      backgroundColor: 'var(--mantine-color-dark-8)',
                      border: '1px solid var(--mantine-color-dark-5)',
                    },
                  }}
                />
                <PasswordInput
                  label={t('login.password')}
                  value={password}
                  onChange={(e) => setPassword(e.currentTarget.value)}
                  required
                  styles={{
                    input: {
                      backgroundColor: 'var(--mantine-color-dark-8)',
                      border: '1px solid var(--mantine-color-dark-5)',
                    },
                  }}
                />
                <Button
                  type="submit"
                  loading={loading}
                  fullWidth
                  size="md"
                  radius="xl"
                  variant="filled"
                  color="cyan"
                  mt="xs"
                >
                  {loading ? t('login.loading') : t('login.submit')}
                </Button>
              </Stack>
            </form>
          </Paper>
        </Stack>
      </Center>
    </Box>
  );
}
