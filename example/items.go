package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show items",
		RunE:  showFunc,
	}
	return cmd
}

func showFunc(cmd *cobra.Command, args []string) error {
	fmt.Printf("----- items start ---\n")
	for _, item := range items {
		fmt.Printf("* %s\n", item)
	}
	fmt.Printf("----- items end -----\n")
	return nil
}

func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an item",
	}

	cmd.AddCommand(newAddAppleCommand())
	cmd.AddCommand(newAddMelonCommand())

	cmd.PersistentFlags().IntP("count", "n", 1, "Number of items to add")
	return cmd
}

func newAddAppleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apple",
		Short: "Add apple",
		RunE:  addAppleFunc,
	}

	cmd.Flags().StringP("color", "c", "red", "Apple color")
	return cmd
}

func addAppleFunc(cmd *cobra.Command, args []string) error {
	color, _ := cmd.Flags().GetString("color")
	count, _ := cmd.Flags().GetInt("count")
	for i := 0; i < count; i++ {
		items = append(items, color+" apple")
	}
	return nil
}

func newAddMelonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "melon",
		Short: "Add melon",
		RunE:  addMelonFunc,
	}

	cmd.Flags().IntP("size", "s", 3, "Melon size in kilograms")
	return cmd
}

func addMelonFunc(cmd *cobra.Command, args []string) error {
	size, _ := cmd.Flags().GetInt("size")
	count, _ := cmd.Flags().GetInt("count")
	for i := 0; i < count; i++ {
		items = append(items, fmt.Sprintf("%dkg melon", size))
	}
	return nil
}
