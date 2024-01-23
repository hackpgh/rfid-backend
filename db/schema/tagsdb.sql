-- Create Members Table
CREATE TABLE IF NOT EXISTS members (
    contact_id INTEGER PRIMARY KEY,
    tag_id INTEGER NOT NULL,
    membership_level INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_members_tag_id ON members(tag_id);

-- Create SafetyTrainings Table
CREATE TABLE IF NOT EXISTS trainings (
    training_name TEXT PRIMARY KEY
);

-- Create SafetyTrainingMembersLink Table
CREATE TABLE IF NOT EXISTS members_trainings_link (
    tag_id INTEGER NOT NULL,
    training_name TEXT NOT NULL,
    FOREIGN KEY (tag_id) REFERENCES members(tag_id),
    FOREIGN KEY (training_name) REFERENCES trainings(training_name),
    UNIQUE (tag_id, training_name)
);
