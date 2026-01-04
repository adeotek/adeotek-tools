import { createClient } from '@/lib/supabase/server';

/**
 * Check if a record has any references from other records
 * This prevents deletion of records that are being used as relations
 */
export async function hasRecordReferences(recordId: string): Promise<boolean> {
  const supabase = await createClient();

  const { data, error } = await supabase
    .from('values')
    .select('id')
    .eq('value_relation', recordId)
    .limit(1);

  if (error) {
    console.error('Error checking record references:', error);
    throw new Error('Failed to check record references');
  }

  return (data?.length ?? 0) > 0;
}

/**
 * Get all records that reference a given record
 * Useful for showing which records are blocking deletion
 */
export async function getRecordReferences(recordId: string): Promise<{
  recordId: string;
  entityName: string;
  attributeName: string;
}[]> {
  const supabase = await createClient();

  const { data, error } = await supabase
    .from('values')
    .select(`
      id,
      record_id,
      attribute_id,
      attributes (
        name,
        display_name,
        entity_id,
        entities (
          name,
          display_name
        )
      )
    `)
    .eq('value_relation', recordId);

  if (error) {
    console.error('Error fetching record references:', error);
    throw new Error('Failed to fetch record references');
  }

  // Type for the nested query result
  type ReferenceData = {
    record_id: string;
    attributes?: {
      display_name?: string;
      entities?: {
        display_name?: string;
      };
    };
  };

  return (data ?? []).map((ref: ReferenceData) => ({
    recordId: ref.record_id,
    entityName: ref.attributes?.entities?.display_name || 'Unknown',
    attributeName: ref.attributes?.display_name || 'Unknown',
  }));
}

/**
 * Validate that a relation value exists
 */
export async function validateRelation(
  attributeId: string,
  relationRecordId: string
): Promise<{ valid: boolean; error?: string }> {
  const supabase = await createClient();

  // First, get the attribute to find the referenced entity
  const { data: attribute, error: attrError } = await supabase
    .from('attributes')
    .select('references_entity_id')
    .eq('id', attributeId)
    .single();

  if (attrError || !attribute) {
    return {
      valid: false,
      error: 'Invalid attribute',
    };
  }

  // Check if the referenced record exists and is of the correct entity type
  const { data: record, error: recordError } = await supabase
    .from('records')
    .select('id, entity_id')
    .eq('id', relationRecordId)
    .single();

  if (recordError || !record) {
    return {
      valid: false,
      error: 'Referenced record does not exist',
    };
  }

  // Type assertion to help TypeScript
  const typedRecord = record as { id: string; entity_id: string };
  const typedAttribute = attribute as { references_entity_id: string | null };

  if (typedRecord.entity_id !== typedAttribute.references_entity_id) {
    return {
      valid: false,
      error: 'Referenced record is of incorrect entity type',
    };
  }

  return { valid: true };
}
