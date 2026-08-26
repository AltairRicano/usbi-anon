import { Button } from '../../../shared/components/ui/Button';

export function LevelActions({ 
  onSave, onCancel, loading, isValid, isEditing 
}: { 
  onSave: () => void; 
  onCancel: () => void; 
  loading: boolean; 
  isValid: boolean;
  isEditing: boolean;
}) {
  return (
    <div className="flex gap-2 mt-6">
      <Button
        variant="primary"
        onClick={onSave}
        disabled={loading || !isValid}
      >
        {isEditing ? 'Guardar Cambios' : 'Crear Nivel'}
      </Button>
      <Button variant="outline" onClick={onCancel} disabled={loading}>
        Cancelar
      </Button>
    </div>
  );
}
