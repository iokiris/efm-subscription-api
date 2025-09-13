    CREATE EXTENSION IF NOT EXISTS "pgcrypto";

    CREATE TABLE IF NOT EXISTS subscriptions (
         id BIGSERIAL PRIMARY KEY,
         service_name TEXT NOT NULL,
         price INT NOT NULL,
         user_id UUID NOT NULL,
         start_date DATE NOT NULL,
         end_date DATE,
         created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
         updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );

    CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
    CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name);
    CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date ON subscriptions(start_date);
    CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions(end_date);