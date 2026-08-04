-- 182: Add display_name column to team_members table
-- Allows inviter to set a display name for the team member, which takes
-- priority over the user's username when rendering the member list.

ALTER TABLE team_members
    ADD COLUMN IF NOT EXISTS display_name VARCHAR(100) DEFAULT '';
