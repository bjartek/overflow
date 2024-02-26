package overflow

// run a script with the given code/filanem an options
func (o *OverflowState) FlixScript(filename string, opts ...OverflowInteractionOption) *OverflowScriptResult {
	interaction := o.BuildInteraction(filename, "flix", opts...)

	result := interaction.runScript()

	if interaction.PrintOptions != nil && !interaction.NoLog {
		result.Print()
	}
	if o.StopOnError && result.Err != nil {
		result.PrintArguments(nil)
		panic(result.Err)
	}
	return result
}

// compose interactionOptions into a new Script function
func (o *OverflowState) FlixScriptFN(outerOpts ...OverflowInteractionOption) OverflowScriptFunction {
	return func(filename string, opts ...OverflowInteractionOption) *OverflowScriptResult {
		outerOpts = append(outerOpts, opts...)
		return o.FlixScript(filename, outerOpts...)
	}
}

// compose fileName and interactionOptions into a new Script function
func (o *OverflowState) FlixScriptFileNameFN(filename string, outerOpts ...OverflowInteractionOption) OverflowScriptOptsFunction {
	return func(opts ...OverflowInteractionOption) *OverflowScriptResult {
		outerOpts = append(outerOpts, opts...)
		return o.FlixScript(filename, outerOpts...)
	}
}

// If you store this in a struct and add arguments to it it will not reset between calls
func (o *OverflowState) FlixTxFN(outerOpts ...OverflowInteractionOption) OverflowTransactionFunction {
	return func(filename string, opts ...OverflowInteractionOption) *OverflowResult {
		// outer has to be first since we need to be able to overwrite
		opts = append(outerOpts, opts...)
		return o.FlixTx(filename, opts...)
	}
}

func (o *OverflowState) FlixTxFileNameFN(filename string, outerOpts ...OverflowInteractionOption) OverflowTransactionOptsFunction {
	return func(opts ...OverflowInteractionOption) *OverflowResult {
		// outer has to be first since we need to be able to overwrite
		opts = append(outerOpts, opts...)
		return o.FlixTx(filename, opts...)
	}
}

// run a flix transaction with the given code/filanem an options
func (o *OverflowState) FlixTx(filename string, opts ...OverflowInteractionOption) *OverflowResult {
	interaction := o.BuildInteraction(filename, "flix", opts...)

	return o.sendTx(interaction)
}
