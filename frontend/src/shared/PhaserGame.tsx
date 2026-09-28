import { forwardRef, useLayoutEffect, useRef } from 'react';
import Phaser from 'phaser';

export interface IRefPhaserGame {
    game: Phaser.Game | null;
    scene: Phaser.Scene | null;
}

export interface PhaserGameProps {
    config: Phaser.Types.Core.GameConfig;
    onGameReady?: (game: Phaser.Game) => void;
    /** Accessible name for the game canvas region (audit finding C5). */
    ariaLabel?: string;
}

export const PhaserGame = forwardRef<IRefPhaserGame, PhaserGameProps>(({ config, onGameReady, ariaLabel }, ref) => {
    const gameContainer = useRef<HTMLDivElement>(null);
    // Tracks the live Phaser.Game instance itself (not DOM state) so React 18
    // StrictMode's synchronous mount->cleanup->mount in dev never races against
    // Phaser's own (not guaranteed synchronous) canvas teardown/creation.
    const gameInstance = useRef<Phaser.Game | null>(null);

    useLayoutEffect(() => {
        if (!gameContainer.current) return;

        // Prevent multiple game instances during React StrictMode development
        if (gameInstance.current) return;

        const game = new Phaser.Game({ ...config, parent: gameContainer.current });
        gameInstance.current = game;

        if (typeof ref === 'function') {
            ref({ game, scene: null });
        } else if (ref) {
            ref.current = { game, scene: null };
        }

        // Call onGameReady when Phaser finishes booting
        game.events.once('ready', () => {
            if (onGameReady) {
                onGameReady(game);
            }
        });

        // Setup a global event listener to keep the active scene reference updated
        game.events.on('step', () => {
            const activeScene = game.scene.getScenes(true)[0];
            if (activeScene) {
                if (typeof ref === 'function') {
                    ref({ game, scene: activeScene });
                } else if (ref) {
                    ref.current = { game, scene: activeScene };
                }
            }
        });

        return () => {
            if (gameInstance.current) {
                gameInstance.current.destroy(true);
                gameInstance.current = null;
            }
            if (typeof ref === 'function') {
                ref({ game: null, scene: null });
            } else if (ref) {
                ref.current = { game: null, scene: null };
            }
        };
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return (
        <div
            ref={gameContainer}
            className="phaser-game-container"
            style={{ width: '100%', height: '100%' }}
            role="group"
            aria-label={ariaLabel ?? 'Área de juego interactivo'}
        />
    );
});
