INSERT INTO services (business_id, name, description, duration_minutes, price)
VALUES

-- Castellanos Barber Shop
((SELECT id FROM businesses WHERE slug = 'castellanos-barber-shop'), 'Classic Haircut', 'Traditional scissor and clipper cut.', 30, 25.00),
((SELECT id FROM businesses WHERE slug = 'castellanos-barber-shop'), 'Skin Fade', 'Clean skin fade with detailed blending.', 45, 35.00),
((SELECT id FROM businesses WHERE slug = 'castellanos-barber-shop'), 'Beard Trim', 'Beard shaping and line-up.', 20, 15.00),
((SELECT id FROM businesses WHERE slug = 'castellanos-barber-shop'), 'Haircut + Beard Combo', 'Full haircut with beard grooming.', 60, 45.00),

-- Downtown Fade Studio
((SELECT id FROM businesses WHERE slug = 'downtown-fade-studio'), 'Fade Cut', 'Modern fade with styling.', 40, 30.00),
((SELECT id FROM businesses WHERE slug = 'downtown-fade-studio'), 'Kids Haircut', 'Haircut for kids under 12.', 25, 20.00),
((SELECT id FROM businesses WHERE slug = 'downtown-fade-studio'), 'Line Up', 'Sharp edge-up and clean finish.', 15, 12.00),
((SELECT id FROM businesses WHERE slug = 'downtown-fade-studio'), 'Hair Wash & Style', 'Wash and professional styling.', 30, 22.00),

-- Elite Cuts
((SELECT id FROM businesses WHERE slug = 'elite-cuts'), 'Executive Cut', 'Premium cut with hot towel finish.', 50, 50.00),
((SELECT id FROM businesses WHERE slug = 'elite-cuts'), 'Hot Towel Shave', 'Classic straight razor shave experience.', 45, 40.00),
((SELECT id FROM businesses WHERE slug = 'elite-cuts'), 'Premium Package', 'Cut, shave, and styling.', 75, 75.00),
((SELECT id FROM businesses WHERE slug = 'elite-cuts'), 'Hair & Beard Sculpt', 'Precision grooming for hair and beard.', 60, 55.00),

-- The Urban Razor
((SELECT id FROM businesses WHERE slug = 'the-urban-razor'), 'Straight Razor Shave', 'Traditional razor shave with hot towel.', 40, 35.00),
((SELECT id FROM businesses WHERE slug = 'the-urban-razor'), 'Urban Fade', 'Clean fade with razor detailing.', 45, 38.00),
((SELECT id FROM businesses WHERE slug = 'the-urban-razor'), 'Buzz Cut', 'Simple and sharp buzz cut.', 20, 18.00),
((SELECT id FROM businesses WHERE slug = 'the-urban-razor'), 'Full Grooming Service', 'Haircut, beard, and styling.', 70, 65.00);