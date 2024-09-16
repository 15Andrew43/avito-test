DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tender_status') THEN
        CREATE TYPE tender_status AS ENUM (
            'CREATED',
            'PUBLISHED',
            'CLOSED'
        );
    END IF;
END $$;



CREATE TABLE IF NOT EXISTS tender (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    service_type VARCHAR(50),
    status tender_status DEFAULT 'CREATED',
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    creator_id UUID REFERENCES employee(id) ON DELETE SET NULL,
    version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS tender_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    service_type VARCHAR(50),
    status tender_status,
    organization_id UUID,
    creator_id UUID,
    version INT,
    updated_at TIMESTAMP
);




CREATE OR REPLACE FUNCTION save_tender_version()
RETURNS TRIGGER AS $$
BEGIN

    INSERT INTO tender_history (tender_id, name, description, service_type, status, organization_id, creator_id, version, updated_at)
    SELECT OLD.id, OLD.name, OLD.description, OLD.service_type, OLD.status, OLD.organization_id, OLD.creator_id, OLD.version, OLD.updated_at;


    IF TG_OP = 'UPDATE' THEN
        NEW.version := OLD.version + 1;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER tender_update_trigger
BEFORE UPDATE ON tender
FOR EACH ROW
EXECUTE FUNCTION save_tender_version();


DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'bid_author_type') THEN
        CREATE TYPE bid_author_type AS ENUM (
            'User',
            'Organization'
        );
    END IF;
END $$;


DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'bid_status') THEN
        CREATE TYPE bid_status AS ENUM (
            'CREATED',
            'PUBLISHED',
            'CANCELED'
        );
    END IF;
END $$;



CREATE TABLE IF NOT EXISTS bid (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID REFERENCES employee(id) ON DELETE CASCADE,
    author_type bid_author_type,
    description TEXT,
    status bid_status DEFAULT 'CREATED',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);



DROP TABLE IF EXISTS bid_review;

CREATE TABLE IF NOT EXISTS bid_review (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bid_id UUID REFERENCES bid(id) ON DELETE CASCADE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);




