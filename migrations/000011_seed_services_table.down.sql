DELETE FROM services
WHERE business_id IN (
    SELECT id FROM businesses
    WHERE slug IN (
        'castellanos-barber-shop',
        'downtown-fade-studio',
        'elite-cuts',
        'the-urban-razor'
    )
);