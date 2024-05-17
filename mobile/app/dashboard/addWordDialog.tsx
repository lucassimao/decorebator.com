import * as wordlistsApi from '@/api/wordlists';
import { useMutation } from '@tanstack/react-query';
import * as React from 'react';
import { Controller, useForm } from 'react-hook-form';
import { StyleSheet } from 'react-native';
import { ActivityIndicator, Button, Dialog, HelperText, Portal, Snackbar, TextInput, useTheme } from 'react-native-paper';

type Props = {
    // success is true if word was created
    onDismiss: (success?: boolean) => void
    wordlistId: number
}

const AddWordDialog = ({ onDismiss, wordlistId }: Props) => {
    const theme = useTheme();

    const { control, handleSubmit, formState: { errors } } = useForm({
        defaultValues: {
            name: '',
        }
    },)

    const [error, setError] = React.useState<any>(null)

    const { mutate: addWord, isPending } = useMutation<void, Error, { name: string }>({
        mutationFn: ({ name }) => wordlistsApi.addWord({ wordlistId, name }),
        onError: (error) => {
            setError(error)
        },
        onSuccess: () => {
            onDismiss(true)
        },
    })


    const onSubmit = (data: any) => addWord({ name: data.name })

    return (
        <Portal>
            <Snackbar
                visible={!!error}
                onDismiss={() => setError(null)}
                action={{
                    label: 'Hide',
                }}>
                {error?.message}
            </Snackbar>

            <Dialog visible={true} onDismiss={onDismiss} dismissable dismissableBackButton>
                <Dialog.Title style={styles.title}>Add Word</Dialog.Title>
                <Dialog.Content>

                    {isPending ? <ActivityIndicator size={'large'} animating={true} theme={theme} /> :
                        <Controller
                            control={control}
                            rules={{
                                required: true,
                            }}
                            render={({ field: { onChange, onBlur, value } }) => (
                                <TextInput
                                    label="Word"
                                    mode='outlined'
                                    // style={styles.inputs}
                                    onBlur={onBlur}
                                    onChangeText={onChange}
                                    value={value}
                                    error={!!errors.name}
                                />
                            )}
                            name="name"
                        />
                    }
                    {errors.name && <HelperText type="error">
                        Required
                    </HelperText>
                    }
                </Dialog.Content>
                <Dialog.Actions>
                    <Button onPress={() => onDismiss(false)}>Cancel</Button>
                    <Button onPress={handleSubmit(onSubmit)}>Add</Button>
                </Dialog.Actions>
            </Dialog>
        </Portal>
    );
};

const styles = StyleSheet.create({
    title: {
        textAlign: 'center',
    },
})

export default AddWordDialog;