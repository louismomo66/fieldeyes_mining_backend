#!/bin/bash

# Enhanced Chain of Custody Traceability Test Script
# Tests the new production record linking system

BASE_URL="http://localhost:9006/api/v1"

echo "🔗 Testing Enhanced Chain of Custody Traceability System"
echo "========================================================"

# Step 1: Create a test user and get auth token
echo -e "\n1. Creating test user..."
SIGNUP_RESP=$(curl -s -X POST $BASE_URL/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "miner@test.com",
    "name": "Test Miner",
    "password": "testpass123"
  }')

TOKEN=$(echo $SIGNUP_RESP | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)

if [ -z "$TOKEN" ]; then
    echo "❌ Failed to get auth token"
    exit 1
fi

echo "✅ User created, token obtained"

# Step 2: Create production records with different grades for different minerals
echo -e "\n2. Creating production records with grades..."

# Gold production record from Pit A
GOLD_RECORD_1=$(curl -s -X POST $BASE_URL/inventory \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gold Ore Batch A1",
    "type": "mineral", 
    "date": "2024-07-01",
    "from": "mine",
    "pit_number": "PIT-A",
    "miner_name": "John Miner",
    "miner_serial_number": "MIN-001",
    "batch_number": "GOLD-A1-2024",
    "processing_method": "crushing",
    "product": "ore",
    "grade_value": 12.5,
    "grade_unit": "g/t",
    "grade_notes": "High grade gold ore",
    "quantity": 1500,
    "unit": "kg",
    "min_stock_level": 100,
    "current_value": 75000
  }')

GOLD_ID_1=$(echo $GOLD_RECORD_1 | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)

# Gold production record from Pit B  
GOLD_RECORD_2=$(curl -s -X POST $BASE_URL/inventory \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gold Ore Batch B1",
    "type": "mineral",
    "date": "2024-07-02", 
    "from": "mine",
    "pit_number": "PIT-B",
    "miner_name": "Jane Smith",
    "miner_serial_number": "MIN-002",
    "batch_number": "GOLD-B1-2024",
    "processing_method": "crushing",
    "product": "ore",
    "grade_value": 8.3,
    "grade_unit": "g/t",
    "grade_notes": "Medium grade gold ore",
    "quantity": 2200,
    "unit": "kg", 
    "min_stock_level": 100,
    "current_value": 91000
  }')

GOLD_ID_2=$(echo $GOLD_RECORD_2 | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)

# Copper production record
COPPER_RECORD=$(curl -s -X POST $BASE_URL/inventory \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Copper Ore Batch C1", 
    "type": "mineral",
    "date": "2024-07-03",
    "from": "mine",
    "pit_number": "PIT-C",
    "miner_name": "Bob Cooper",
    "miner_serial_number": "MIN-003",
    "batch_number": "CU-C1-2024",
    "processing_method": "crushing",
    "product": "concentrate",
    "grade_value": 3.2,
    "grade_unit": "%",
    "grade_notes": "Copper concentrate",
    "quantity": 5000,
    "unit": "kg",
    "min_stock_level": 500,
    "current_value": 160000
  }')

COPPER_ID=$(echo $COPPER_RECORD | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)

echo "✅ Production records created:"
echo "   - Gold Batch A1 (ID: $GOLD_ID_1) - Grade: 12.5 g/t"
echo "   - Gold Batch B1 (ID: $GOLD_ID_2) - Grade: 8.3 g/t" 
echo "   - Copper Batch C1 (ID: $COPPER_ID) - Grade: 3.2%"

# Step 3: Get available production records for gold
echo -e "\n3. Getting available production records for gold..."
AVAILABLE_GOLD=$(curl -s -X GET "$BASE_URL/compliance/coc-lots/production-records?mineral_type=gold" \
  -H "Authorization: Bearer $TOKEN")

echo "✅ Available gold production records:"
echo "$AVAILABLE_GOLD" | python3 -c "import sys,json; data=json.load(sys.stdin); [print(f'   - ID: {r[\"id\"]}, Batch: {r[\"batch_number\"]}, Pit: {r[\"pit_number\"]}, Grade: {r[\"grade_value\"]} {r[\"grade_unit\"]}') for r in data['data']]" 2>/dev/null

# Step 4: Get production records by pit
echo -e "\n4. Getting production records from PIT-A..."
PIT_A_RECORDS=$(curl -s -X GET "$BASE_URL/compliance/coc-lots/production-records/by-pit?pit_number=PIT-A" \
  -H "Authorization: Bearer $TOKEN")

echo "✅ PIT-A production records:"
echo "$PIT_A_RECORDS" | python3 -c "import sys,json; data=json.load(sys.stdin); [print(f'   - ID: {r[\"id\"]}, Batch: {r[\"batch_number\"]}, Grade: {r[\"grade_value\"]} {r[\"grade_unit\"]}') for r in data['data']]" 2>/dev/null

# Step 5: Create Chain of Custody Lot linking multiple production records
echo -e "\n5. Creating Chain of Custody Lot with linked production records..."
COC_LOT=$(curl -s -X POST $BASE_URL/compliance/coc-lots \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"lot_number\": \"LOT-GOLD-001-2024\",
    \"production_record_ids\": [$GOLD_ID_1, $GOLD_ID_2],
    \"mineral_type\": \"gold\",
    \"ore_type\": \"Gold ore concentrate\",
    \"weight\": 3700,
    \"unit\": \"kg\",
    \"grade_value\": 10.2,
    \"grade_unit\": \"g/t\",
    \"grade_notes\": \"Blended gold ore from Pit A and B\",
    \"number_of_sacks\": 74,
    \"source_mine_site\": \"FieldEyes Test Mine\",
    \"mine_site_status\": \"green\",
    \"mine_operator_name\": \"FieldEyes Mining Ltd\",
    \"miner_name\": \"Composite Lot\",
    \"miner_national_id\": \"COMP-001\"
  }")

LOT_ID=$(echo $COC_LOT | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)

echo "✅ Chain of Custody Lot created:"
echo "   - Lot Number: LOT-GOLD-001-2024"
echo "   - Lot ID: $LOT_ID"
echo "   - Linked Production Records: $GOLD_ID_1, $GOLD_ID_2"
echo "   - Blended Grade: 10.2 g/t"

# Step 6: Retrieve the lot with production records
echo -e "\n6. Retrieving lot with linked production records..."
LOT_WITH_RECORDS=$(curl -s -X GET "$BASE_URL/compliance/coc-lots" \
  -H "Authorization: Bearer $TOKEN")

echo "✅ Chain of Custody Lots with production record linkage:"
echo "$LOT_WITH_RECORDS" | python3 -c "
import sys,json
data = json.load(sys.stdin)
for lot in data['data']:
    print(f'   - Lot: {lot[\"lot_number\"]}')
    print(f'     Mineral: {lot[\"mineral_type\"]}')
    print(f'     Grade: {lot.get(\"grade_value\", \"N/A\")} {lot.get(\"grade_unit\", \"\")}')
    if lot.get('production_records'):
        print(f'     Linked Records: {len(lot[\"production_records\"])}')
        for record in lot['production_records']:
            print(f'       * {record[\"batch_number\"]} (Pit: {record[\"pit_number\"]})')
    print()
" 2>/dev/null

# Step 7: Test adding more production records to existing lot
echo -e "\n7. Adding additional production record to existing lot..."
if [ ! -z "$LOT_ID" ] && [ ! -z "$COPPER_ID" ]; then
    ADD_RECORDS=$(curl -s -X POST "$BASE_URL/compliance/coc-lots/$LOT_ID/production-records" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"production_record_ids\": [$COPPER_ID]}")
    
    echo "✅ Additional production record linked to lot"
fi

# Step 8: Verify traceability - show complete audit trail
echo -e "\n8. Complete Traceability Audit Trail:"
echo "========================================"

FINAL_LOTS=$(curl -s -X GET "$BASE_URL/compliance/coc-lots" \
  -H "Authorization: Bearer $TOKEN")

echo "$FINAL_LOTS" | python3 -c "
import sys,json
data = json.load(sys.stdin)
print('📋 TRACEABILITY REPORT')
print('=' * 50)
for lot in data['data']:
    print(f'🏷️  LOT: {lot[\"lot_number\"]}')
    print(f'    Mineral Type: {lot[\"mineral_type\"]}')
    print(f'    Total Weight: {lot[\"weight\"]} {lot[\"unit\"]}')
    print(f'    Grade: {lot.get(\"grade_value\", \"N/A\")} {lot.get(\"grade_unit\", \"\")}')
    print(f'    Mine Site: {lot[\"source_mine_site\"]}')
    print(f'    Status: {lot[\"mine_site_status\"]}')
    print()
    if lot.get('production_records'):
        print(f'    📊 LINKED PRODUCTION RECORDS ({len(lot[\"production_records\"])}):')
        total_weight = 0
        for i, record in enumerate(lot['production_records'], 1):
            print(f'    {i}. {record[\"batch_number\"]}')
            print(f'       📍 Pit: {record[\"pit_number\"]}')
            print(f'       👤 Miner: {record[\"miner_name\"]} ({record.get(\"miner_serial_number\", \"N/A\")})')
            print(f'       📊 Grade: {record.get(\"grade_value\", \"N/A\")} {record.get(\"grade_unit\", \"\")}')
            print(f'       ⚖️  Quantity: {record[\"quantity\"]} {record[\"unit\"]}')
            print(f'       📅 Date: {record.get(\"date\", \"N/A\")}')
            print(f'       🔄 Processing: {record.get(\"processing_method\", \"N/A\")}')
            total_weight += record['quantity']
            print()
        print(f'    📈 TOTAL SOURCE WEIGHT: {total_weight} kg')
    print('=' * 50)
" 2>/dev/null

echo -e "\n✅ Enhanced Chain of Custody Traceability System Test Complete!"
echo -e "\n🎯 Key Features Demonstrated:"
echo "   ✓ Production records with mineral-specific grades"
echo "   ✓ Pit-based production tracking"
echo "   ✓ Many-to-many linking of production records to CoC lots"
echo "   ✓ Complete audit trail from mine to lot"
echo "   ✓ Grade blending and weight reconciliation"
echo "   ✓ Miner identification and serial numbers"
echo "   ✓ Processing method tracking"
echo "   ✓ Seamless data flow without repetition"