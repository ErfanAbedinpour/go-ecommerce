ALTER TABLE store_settings
    ADD COLUMN IF NOT EXISTS checkout JSONB NOT NULL DEFAULT '{
        "min_order_toman": 100000,
        "payment_methods": ["online", "cod"],
        "cod_enabled": true,
        "cod_cities": ["تهران", "کرج", "Tehran", "Karaj"],
        "currency_label": "تومان"
    }';
