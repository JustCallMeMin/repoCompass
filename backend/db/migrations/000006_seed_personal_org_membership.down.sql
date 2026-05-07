-- Remove only the local-dev membership seeded by this migration.

DELETE FROM organization_memberships
WHERE organization_id = '00000000-0000-0000-0000-000000000000'
  AND user_id = 'mock_user';
